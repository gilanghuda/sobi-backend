package controllers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
)

func GetAllEducations(c *fiber.Ctx) error {
	db := database.DB
	if db == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "database not initialized"})
	}

	eds, err := queries.GetAllEducations(db)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(eds)
}
