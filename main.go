package main

import (
	"log"

	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gilanghuda/sobi-backend/pkg/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	_, err = database.InitDB()
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	routes.RegisterUserRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
