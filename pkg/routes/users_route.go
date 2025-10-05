package routes

import (
	"github.com/gilanghuda/sobi-backend/app/controllers"
	"github.com/gilanghuda/sobi-backend/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {
	user := app.Group("/user", middleware.JWTProtected())
	user.Get("/profile", controllers.UserProfile)
	user.Put("/profile", controllers.UpdateUser)
	user.Delete("/profile", controllers.DeleteUser)
	user.Post("/logout", controllers.UserLogout)

	app.Post("/signup", controllers.UserSignUp)
	app.Post("/signin", controllers.UserSignIn)
	app.Post("/signin/google", controllers.UserSignInWithGoogle)
	app.Post("/verify-otp", controllers.UserVerifyOTP)
	app.Post("/refresh-token", controllers.RefreshToken)
	app.Get("/get-ahli", controllers.GetAhliWithDetails)
	app.Get("/user/:id", controllers.GetUserByID)
	app.Post("/ahli", controllers.PromoteToAhli)

}
