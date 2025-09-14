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
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
