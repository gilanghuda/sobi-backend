package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type UserGoal struct {
	ID            uuid.UUID `json:"id" db:"id"`
	UserID        uuid.UUID `json:"user_id" db:"user_id"`
	GoalCategory  string    `json:"goal_category" db:"goal_category"`
	Status        string    `json:"status" db:"status"`
	CurrentDay    int       `json:"current_day" db:"current_day"`
	StartDate     time.Time `json:"start_date" db:"start_date"`
	TargetEndDate time.Time `json:"target_end_date" db:"target_end_date"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type CreateUserGoalRequest struct {
	GoalCategory  string `json:"goal_category,omitempty" validate:"required"`
	StartDate     string `json:"start_date,omitempty" validate:"required"`
	TargetEndDate string `json:"target_end_date,omitempty" validate:"required"`
}

type CreateGoalSummaryRequest struct {
	UserGoalID  string   `json:"user_goal_id,omitempty"`
	Reflection  []string `json:"reflection,omitempty"`
	SelfChanges []string `json:"self_changes,omitempty"`
}

type GoalSummary struct {
	ID                   uuid.UUID       `json:"id" db:"id"`
	UserGoalID           uuid.UUID       `json:"user_goal_id" db:"user_goal_id"`
	UserID               uuid.UUID       `json:"user_id" db:"user_id"`
	GoalCategory         string          `json:"goal_category" db:"goal_category"`
	TotalDays            int             `json:"total_days" db:"total_days"`
	DaysCompleted        int             `json:"days_completed" db:"days_completed"`
	TotalMissions        int             `json:"total_missions" db:"total_missions"`
	MissionsCompleted    int             `json:"missions_completed" db:"missions_completed"`
	TotalTasks           int             `json:"total_tasks" db:"total_tasks"`
	TasksCompleted       int             `json:"tasks_completed" db:"tasks_completed"`
	CompletionPercentage float64         `json:"completion_percentage" db:"completion_percentage"`
	ReflectionText       string          `json:"reflection" db:"reflection"`
	SelfChanges          json.RawMessage `json:"self_changes" db:"self_changes"`
	StartDate            time.Time       `json:"start_date" db:"start_date"`
	EndDate              time.Time       `json:"end_date" db:"end_date"`
	CreatedAt            time.Time       `json:"created_at" db:"created_at"`
}
