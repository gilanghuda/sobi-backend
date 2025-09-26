package models

import "time"

// Education represents the educations table
type Education struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Subtitle    *string   `json:"subtitle,omitempty"`
	VideoURL    *string   `json:"video_url,omitempty"`
	Duration    *string   `json:"duration,omitempty"`
	Author      *string   `json:"author,omitempty"`
	Description *string   `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}
