package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `json:"id" db:"uid"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PhoneNumber  string    `json:"phone_number,omitempty"`
	Gender       string    `json:"gender,omitempty"`
	Avatar       string    `json:"avatar,omitempty"`
	PasswordHash string    `json:"-"`
	Verified     bool      `json:"verified"`
	OTP          string    `json:"-"`
	UserRole     string    `json:"user_role"`

	Price    float64 `json:"price,omitempty"`
	Category string  `json:"category,omitempty"`
	OpenTime string  `json:"open_time,omitempty"`
	Rating   float64 `json:"rating,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PromoteAhliRequest struct {
	UserID   uuid.UUID `json:"user_id,omitempty"`
	Price    float64   `json:"price"`
	Category string    `json:"category"`
	OpenTime string    `json:"open_time"`
	Rating   float64   `json:"rating,omitempty"`
}
