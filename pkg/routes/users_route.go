package routes

import (
	"github.com/gilanghuda/sobi-backend/app/controllers"
	"github.com/gilanghuda/sobi-backend/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterUserRoutes(app *fiber.App) {

	app.Post("/signup", controllers.UserSignUp)
	app.Post("/signin", controllers.UserSignIn)
	app.Post("/verify-otp", controllers.UserVerifyOTP)
	app.Post("/refresh-token", controllers.RefreshToken)

	user := app.Group("/user", middleware.JWTProtected())
	user.Get("/profile", controllers.UserProfile)
	user.Put("/profile", controllers.UpdateUser)
	user.Delete("/profile", controllers.DeleteUser)
	user.Post("/logout", controllers.UserLogout)
}
