package main

import (
	"log"

	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gilanghuda/sobi-backend/pkg/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/gilanghuda/sobi-backend/app/controllers"
)

func main() {
	_ = godotenv.Load()
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %v", err)
	// }

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3001, http://localhost:3002, http://localhost:3003, https://sobi.gilanghuda.my.id",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
	}))

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	_, err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	routes.RegisterUserRoutes(app)
	routes.RegisterGoalsRoutes(app)
	routes.RegisterChatRoutes(app)
	routes.RegisterEducationRoutes(app)
	routes.RegisterTransactionRoutes(app)

	controllers.StartMessageDispatcher()

	log.Fatal(app.Listen(":8000"))
}
