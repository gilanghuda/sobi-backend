package controllers

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gilanghuda/sobi-backend/pkg/utils"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var validate = validator.New()

func UserSignUp(c *fiber.Ctx) error {
	signUp := &models.SignUp{}
	if err := c.BodyParser(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validate.Struct(signUp); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	role := signUp.UserRole
	if role == "" {
		role = utils.RoleUser
	}

	valid := false
	for _, r := range utils.ValidRoles {
		if role == r {
			valid = true
			break
		}
	}
	if !valid {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user role",
		})
	}

	userQueries := queries.UserQueries{DB: database.DB}
	existing, err := userQueries.GetUserByEmail(signUp.Email)
	if err == nil {
		if !existing.Verified {
			otp, err := utils.GenerateOTP(4)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate OTP"})
			}
			if err := userQueries.UpdateOTPByEmail(signUp.Email, otp); err != nil {
				println(err.Error())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update OTP"})
			}
			if err := utils.SendOTPEmail(signUp.Email, otp); err != nil {
				println(
					err.Error(),
				)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP email"})
			}
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "OTP resent to email"})
		}
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "Email already registered"})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(signUp.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to hash password"})
	}

	otp, err := utils.GenerateOTP(4)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate OTP"})
	}

	user := &models.User{
		ID:           uuid.New(),
		Email:        signUp.Email,
		Username:     signUp.Username,
		PasswordHash: string(hashedPassword),
		UserRole:     role,
		PhoneNumber:  signUp.Phone,
		Gender:       "male",
		Avatar:       "1",
		Verified:     false,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		OTP:          otp,
	}

	if err := userQueries.CreateUser(user); err != nil {
		println(err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user"})
	}

	if err := utils.SendOTPEmail(signUp.Email, otp); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to send OTP email"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User registered. OTP sent to email"})
}

func UserVerifyOTP(c *fiber.Ctx) error {
	payload := &models.VerifyOTP{}
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := validate.Struct(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	userQueries := queries.UserQueries{DB: database.DB}
	if err := userQueries.VerifyOTPByEmail(payload.Email, payload.OTP); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Account verified successfully"})
}

func UserSignIn(c *fiber.Ctx) error {
	signIn := &models.SignIn{}
	if err := c.BodyParser(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := validate.Struct(signIn); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	userQueries := queries.UserQueries{DB: database.DB}
	user, err := userQueries.GetUserByEmail(signIn.Email)
	if err != nil {
		println(err.Error())
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	if !user.Verified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Account not verified. Please verify your account before signing in",
		})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(signIn.Password)); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid email or password",
		})
	}

	accessMinutes := 15
	if v := os.Getenv("ACCESS_TOKEN_MINUTES"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			accessMinutes = iv
		}
	}
	refreshHours := 24 * 7
	if v := os.Getenv("REFRESH_TOKEN_HOURS"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			refreshHours = iv
		}
	}

	secret := os.Getenv("JWT_SECRET")
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"email":     user.Email,
		"user_role": user.UserRole,
		"exp":       time.Now().Add(time.Duration(accessMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	rtStr, err := utils.GenerateRandomToken(32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate refresh token"})
	}
	rt := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		Token:     rtStr,
		ExpiresAt: time.Now().Add(time.Duration(refreshHours) * time.Hour),
		Revoked:   false,
		CreatedAt: time.Now(),
	}
	rtQueries := queries.RefreshTokenQueries{DB: database.DB}
	if err := rtQueries.CreateRefreshToken(rt); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to store refresh token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":            "Sign in successful",
		"access_token":       tokenString,
		"expires_in":         accessMinutes * 60,
		"refresh_token":      rtStr,
		"refresh_expires_at": rt.ExpiresAt,
		"user": fiber.Map{
			"id":        user.ID,
			"email":     user.Email,
			"user_role": user.UserRole,
		},
	})
}

func RefreshToken(c *fiber.Ctx) error {
	payload := struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	if err := validate.Struct(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	rtQueries := queries.RefreshTokenQueries{DB: database.DB}
	rt, err := rtQueries.GetRefreshTokenByToken(payload.RefreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid refresh token"})
	}

	if rt.Revoked || time.Now().After(rt.ExpiresAt) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Refresh token expired or revoked"})
	}

	userQueries := queries.UserQueries{DB: database.DB}
	user, err := userQueries.GetUserByID(rt.UserID)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "User not found"})
	}

	accessMinutes := 15
	if v := os.Getenv("ACCESS_TOKEN_MINUTES"); v != "" {
		if iv, err := strconv.Atoi(v); err == nil {
			accessMinutes = iv
		}
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "JWT secret not set"})
	}

	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"email":     user.Email,
		"user_role": user.UserRole,
		"exp":       time.Now().Add(time.Duration(accessMinutes) * time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to generate access token"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"access_token": tokenString, "expires_in": accessMinutes * 60})
}

func UserLogout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Missing or invalid Authorization header"})
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "JWT secret not set"})
	}

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) { return []byte(secret), nil })
	if err != nil || !token.Valid {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid or expired token"})
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token claims"})
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid token payload"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user id in token"})
	}

	body := struct {
		RefreshToken string `json:"refresh_token"`
	}{}
	_ = c.BodyParser(&body)

	rtQueries := queries.RefreshTokenQueries{DB: database.DB}
	if body.RefreshToken != "" {
		if err := rtQueries.RevokeRefreshTokenByToken(body.RefreshToken); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to revoke refresh token"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Refresh token revoked"})
	}

	if err := rtQueries.RevokeRefreshTokensByUser(userID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to revoke refresh tokens for user"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "Logged out"})
}
