package models

import (
	"time"

	"github.com/google/uuid"
)

type Mission struct {
	ID        uuid.UUID `json:"id" db:"id"`
	DayNumber int       `json:"day_number" db:"day_number"`
	Focus     string    `json:"focus" db:"focus"`
	Category  string    `json:"category" db:"category"`
}

type Task struct {
	ID        uuid.UUID `json:"id" db:"id"`
	MissionID uuid.UUID `json:"mission_id" db:"mission_id"`
	Text      string    `json:"text" db:"text"`
}

type TaskProgress struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserGoalID  uuid.UUID  `json:"user_goal_id" db:"user_goal_id"`
	TaskID      uuid.UUID  `json:"task_id" db:"task_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	IsCompleted bool       `json:"is_completed" db:"is_completed"`
	CompletedAt *time.Time `json:"completed_at" db:"completed_at"`
}

type MissionProgress struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	UserGoalID     uuid.UUID  `json:"user_goal_id" db:"user_goal_id"`
	MissionID      uuid.UUID  `json:"mission_id" db:"mission_id"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	IsCompleted    bool       `json:"is_completed" db:"is_completed"`
	CompletedAt    *time.Time `json:"completed_at" db:"completed_at"`
	TotalTasks     int        `json:"total_tasks" db:"total_tasks"`
	CompletedTasks int        `json:"completed_tasks" db:"completed_tasks"`
	CompletionPerc float64    `json:"completion_percentage" db:"completion_percentage"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// View / composite structs used by controllers
type TaskWithProgress struct {
	Task     Task         `json:"task,omitempty"`
	Progress TaskProgress `json:"progress,omitempty"`
}

type MissionWithTasks struct {
	Mission  Mission            `json:"mission,omitempty"`
	Tasks    []TaskWithProgress `json:"tasks,omitempty"`
	Progress MissionProgress    `json:"progress,omitempty"`
}

type GoalDay struct {
	UserGoal UserGoal           `json:"user_goal,omitempty"`
	DayIndex int                `json:"day_index,omitempty"`
	Missions []MissionWithTasks `json:"missions,omitempty"`
}
