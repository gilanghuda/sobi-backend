package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gilanghuda/sobi-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateTransaction(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	p := &models.CreateTransactionRequest{}
	if err := c.BodyParser(p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if p.AhliID == "" || p.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "ahli_id and positive amount are required"})
	}
	ahliID, err := uuid.Parse(p.AhliID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid ahli_id"})
	}

	tx := &models.Transaction{ID: uuid.New(), UserID: userID, AhliID: ahliID, Amount: p.Amount, Status: "pending", CreatedAt: time.Now(), UpdatedAt: time.Now()}

	paymentURL, err := utils.CreateMidtransTransaction(tx.ID.String(), tx.Amount)
	if err != nil {
		return c.Status(http.StatusBadGateway).JSON(fiber.Map{"error": "failed to create midtrans transaction"})
	}
	tx.PaymentURL = paymentURL

	q := queries.TransactionQueries{DB: database.DB}
	if err := q.CreateTransaction(tx); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create transaction"})
	}

	resp := models.CreateTransactionResponse{ID: tx.ID, PaymentURL: tx.PaymentURL}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func GetTransactionByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "missing id"})
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id"})
	}
	q := queries.TransactionQueries{DB: database.DB}
	tx, err := q.GetTransactionByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "transaction not found"})
	}
	return c.Status(fiber.StatusOK).JSON(tx)
}

func MidtransNotification(c *fiber.Ctx) error {
	var payload map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(c.Body())).Decode(&payload); err != nil {
		if err := c.BodyParser(&payload); err != nil {
			return c.SendStatus(http.StatusBadRequest)
		}
	}

	orderID, _ := payload["order_id"].(string)
	if orderID == "" {
		if td, ok := payload["transaction_details"].(map[string]interface{}); ok {
			orderID, _ = td["order_id"].(string)
		}
	}
	if orderID == "" {
		return c.SendStatus(http.StatusBadRequest)
	}

	txStatus, _ := payload["transaction_status"].(string)
	if txStatus == "" {
		if s, ok := payload["status_code"].(string); ok {
			txStatus = s
		}
	}

	localStatus := "pending"
	switch txStatus {
	case "capture", "settlement", "200":
		localStatus = "completed"
	case "pending":
		localStatus = "pending"
	case "deny", "expire", "cancel", "201", "202", "407":
		localStatus = "failed"
	default:
		// keep pending
	}

	id, err := uuid.Parse(orderID)
	if err != nil {
		return c.SendStatus(http.StatusBadRequest)
	}

	q := queries.TransactionQueries{DB: database.DB}
	if err := q.UpdateTransactionStatus(id, localStatus); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "unable to update transaction"})
	}

	return c.SendStatus(http.StatusOK)
}
