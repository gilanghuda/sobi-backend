package models

// HistoryEducation represents a record in history_education
type HistoryEducation struct {
	ID          string `json:"id,omitempty"`
	UserID      string `json:"user_id"`
	EducationID string `json:"education_id"`
}
