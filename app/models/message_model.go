package models

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID `json:"id" db:"id"`
	OwnerID   uuid.UUID `json:"owner_id" db:"owner_id"`
	Category  string    `json:"category" db:"category"`
	Visible   bool      `json:"visible" db:"visible"` // kolumn to mark visibility for matchmaking
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type Message struct {
	ID        uuid.UUID `json:"id" db:"id"`
	RoomID    uuid.UUID `json:"room_id" db:"room_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Text      string    `json:"text" db:"text"`
	Visible   bool      `json:"visible" db:"visible"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateRoomRequest struct {
	Category string `json:"category,omitempty"`
	Visible  bool   `json:"visible,omitempty"`
}

type CreateMessageRequest struct {
	RoomID  string `json:"room_id,omitempty"`
	Text    string `json:"text,omitempty"`
	Visible *bool  `json:"visible,omitempty"`
}

type MatchRequest struct {
	Category string `json:"category,omitempty"`
	Role     string `json:"role,omitempty"`
}
