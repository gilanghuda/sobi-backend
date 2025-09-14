package controllers

import (
	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func UserProfile(c *fiber.Ctx) error {
	claims := c.Locals("user")
	var mapClaims map[string]interface{}

	switch v := claims.(type) {
	case map[string]interface{}:
		mapClaims = v
	case jwt.MapClaims:
		mapClaims = map[string]interface{}(v)
	default:
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token claims",
		})
	}

	userIDStr, ok := mapClaims["user_id"].(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid user id in token",
		})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user id format",
		})
	}

	userQueries := queries.UserQueries{DB: database.DB}
	user, err := userQueries.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	user.PasswordHash = ""

	return c.Status(fiber.StatusOK).JSON(user)
}
