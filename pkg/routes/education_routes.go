package routes

import (
	"github.com/gilanghuda/sobi-backend/app/controllers"

	"github.com/gofiber/fiber/v2"
)

func RegisterEducationRoutes(app *fiber.App) {
	app.Get("/educations", controllers.GetAllEducations)
}
