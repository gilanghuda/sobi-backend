package controllers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gilanghuda/sobi-backend/pkg/utils"
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

func GetEducationDetail(c *fiber.Ctx) error {
	db := database.DB
	if db == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "database not initialized"})
	}

	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}

	edu, err := queries.GetEducationByID(db, id)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if edu == nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{"error": "education not found"})
	}

	auth := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(auth)
	if err == nil {
		history := models.HistoryEducation{
			UserID:      userID.String(),
			EducationID: id,
		}
		_ = queries.InsertHistoryEducation(db, history)
	}

	return c.JSON(edu)
}

func GetUserHistory(c *fiber.Ctx) error {
	db := database.DB
	if db == nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "database not initialized"})
	}

	auth := c.Get("Authorization")
	if auth == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "missing authorization header"})
	}

	userID, err := utils.ExtractUserIDFromHeader(auth)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	eds, err := queries.GetHistoryByUser(db, userID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(eds)
}
