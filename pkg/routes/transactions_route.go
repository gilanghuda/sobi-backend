package routes

import (
	"github.com/gilanghuda/sobi-backend/app/controllers"
	"github.com/gofiber/fiber/v2"
)

func RegisterTransactionRoutes(app *fiber.App) {
	app.Post("/transactions", controllers.CreateTransaction)
	app.Get("/transactions/:id", controllers.GetTransactionByID)
	app.Post("/transactions/notify", controllers.MidtransNotification)
}
