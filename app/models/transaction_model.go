package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	AhliID     uuid.UUID `json:"ahli_id" db:"ahli_id"`
	Amount     int64     `json:"amount" db:"amount"`
	Status     string    `json:"status" db:"status"`
	PaymentURL string    `json:"payment_url,omitempty" db:"payment_url"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

type CreateTransactionRequest struct {
	AhliID string `json:"ahli_id,omitempty"`
	Amount int64  `json:"amount,omitempty"`
}

type CreateTransactionResponse struct {
	ID         uuid.UUID `json:"id"`
	PaymentURL string    `json:"payment_url"`
}
