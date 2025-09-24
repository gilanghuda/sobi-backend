package queries

import (
	"database/sql"
	"errors"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/google/uuid"
)

type MissionsQueries struct {
	DB *sql.DB
}

func (q *MissionsQueries) CreateMission(m *models.Mission) error {
	query := `INSERT INTO missions (id, day_number, focus, category) VALUES ($1, $2, $3, $4)`
	_, err := q.DB.Exec(query, m.ID, m.DayNumber, m.Focus, m.Category)
	if err != nil {
		return errors.New("unable to create mission, DB error")
	}
	return nil
}

func (q *MissionsQueries) CreateTask(t *models.Task) error {
	query := `INSERT INTO tasks (id, mission_id, text) VALUES ($1, $2, $3)`
	_, err := q.DB.Exec(query, t.ID, t.MissionID, t.Text)
	if err != nil {
		return errors.New("unable to create task, DB error")
	}
	return nil
}

func (q *MissionsQueries) GetMissionsByDay(day int) ([]models.Mission, error) {
	var missions []models.Mission
	query := `SELECT id, day_number, focus, category FROM missions WHERE day_number = $1`
	rows, err := q.DB.Query(query, day)
	if err != nil {
		return missions, errors.New("unable to query missions")
	}
	defer rows.Close()
	for rows.Next() {
		var m models.Mission
		if err := rows.Scan(&m.ID, &m.DayNumber, &m.Focus, &m.Category); err != nil {
			return missions, err
		}
		missions = append(missions, m)
	}
	return missions, nil
}

func (q *MissionsQueries) GetMissionsByDayAndCategory(day int, category string) ([]models.Mission, error) {
	var missions []models.Mission
	query := `SELECT id, day_number, focus, category FROM missions WHERE day_number = $1 AND category = $2`
	rows, err := q.DB.Query(query, day, category)
	if err != nil {
		return missions, errors.New("unable to query missions")
	}
	defer rows.Close()
	for rows.Next() {
		var m models.Mission
		if err := rows.Scan(&m.ID, &m.DayNumber, &m.Focus, &m.Category); err != nil {
			return missions, err
		}
		missions = append(missions, m)
	}
	return missions, nil
}

func (q *MissionsQueries) GetTasksByMission(missionID uuid.UUID) ([]models.Task, error) {
	var tasks []models.Task
	query := `SELECT id, mission_id, text FROM tasks WHERE mission_id = $1`
	rows, err := q.DB.Query(query, missionID)
	if err != nil {
		return tasks, errors.New("unable to query tasks")
	}
	defer rows.Close()
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.ID, &t.MissionID, &t.Text); err != nil {
			return tasks, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (q *MissionsQueries) GetTaskProgress(userGoalID, taskID, userID uuid.UUID) (models.TaskProgress, error) {
	p := models.TaskProgress{}
	query := `SELECT id, user_goal_id, task_id, user_id, is_completed, completed_at FROM task_progress WHERE user_goal_id = $1 AND task_id = $2 AND user_id = $3`
	err := q.DB.QueryRow(query, userGoalID, taskID, userID).Scan(&p.ID, &p.UserGoalID, &p.TaskID, &p.UserID, &p.IsCompleted, &p.CompletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, nil
		}
		return p, errors.New("unable to query task progress")
	}
	return p, nil
}

func (q *MissionsQueries) GetMissionProgress(userGoalID, missionID, userID uuid.UUID) (models.MissionProgress, error) {
	mp := models.MissionProgress{}
	query := `SELECT id, user_goal_id, mission_id, user_id, is_completed, completed_at, total_tasks, completed_tasks, completion_percentage, created_at, updated_at FROM mission_progress WHERE user_goal_id = $1 AND mission_id = $2 AND user_id = $3`
	err := q.DB.QueryRow(query, userGoalID, missionID, userID).Scan(&mp.ID, &mp.UserGoalID, &mp.MissionID, &mp.UserID, &mp.IsCompleted, &mp.CompletedAt, &mp.TotalTasks, &mp.CompletedTasks, &mp.CompletionPerc, &mp.CreatedAt, &mp.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return mp, nil
		}
		return mp, errors.New("unable to query mission progress")
	}
	return mp, nil
}

// UpdateTaskProgress executes a single UPDATE statement and returns rows affected
func (q *MissionsQueries) UpdateTaskProgress(userGoalID, taskID, userID uuid.UUID, isCompleted bool) (int64, error) {
	query := `UPDATE task_progress SET is_completed = $1, completed_at = CASE WHEN $1 THEN now() ELSE NULL END WHERE user_goal_id = $2 AND task_id = $3 AND user_id = $4`
	res, err := q.DB.Exec(query, isCompleted, userGoalID, taskID, userID)
	if err != nil {
		return 0, errors.New("unable to update task progress, DB error")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

// InsertTaskProgress inserts a new task_progress record
func (q *MissionsQueries) InsertTaskProgress(id uuid.UUID, userGoalID, taskID, userID uuid.UUID, isCompleted bool) error {
	query := `INSERT INTO task_progress (id, user_goal_id, task_id, user_id, is_completed, completed_at) VALUES ($1, $2, $3, $4, $5, CASE WHEN $5 THEN now() ELSE NULL END)`
	_, err := q.DB.Exec(query, id, userGoalID, taskID, userID, isCompleted)
	if err != nil {
		return errors.New("unable to insert task progress, DB error")
	}
	return nil
}

// CountTotalTasks returns number of tasks for a mission
func (q *MissionsQueries) CountTotalTasks(missionID uuid.UUID) (int, error) {
	var total int
	query := `SELECT count(*) FROM tasks WHERE mission_id = $1`
	if err := q.DB.QueryRow(query, missionID).Scan(&total); err != nil {
		return 0, errors.New("unable to count total tasks")
	}
	return total, nil
}

// CountCompletedTasks returns number of completed tasks for user goal and mission
func (q *MissionsQueries) CountCompletedTasks(userGoalID, missionID, userID uuid.UUID) (int, error) {
	var completed int
	query := `SELECT count(tp.id) FROM task_progress tp JOIN tasks t ON tp.task_id = t.id WHERE tp.user_goal_id = $1 AND t.mission_id = $2 AND tp.user_id = $3 AND tp.is_completed = true`
	if err := q.DB.QueryRow(query, userGoalID, missionID, userID).Scan(&completed); err != nil {
		return 0, errors.New("unable to count completed tasks")
	}
	return completed, nil
}

// UpdateMissionProgress updates existing mission_progress row; returns rows affected
func (q *MissionsQueries) UpdateMissionProgress(userGoalID, missionID, userID uuid.UUID, total, completed int, perc float64, isCompleted bool) (int64, error) {
	query := `UPDATE mission_progress SET total_tasks = $1, completed_tasks = $2, completion_percentage = $3, is_completed = $4, completed_at = CASE WHEN $4 THEN now() ELSE NULL END, updated_at = now() WHERE user_goal_id = $5 AND mission_id = $6 AND user_id = $7`
	res, err := q.DB.Exec(query, total, completed, perc, isCompleted, userGoalID, missionID, userID)
	if err != nil {
		return 0, errors.New("unable to update mission progress, DB error")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

// InsertMissionProgress inserts a new mission_progress row
func (q *MissionsQueries) InsertMissionProgress(id uuid.UUID, userGoalID, missionID, userID uuid.UUID, total, completed int, perc float64, isCompleted bool) error {
	query := `INSERT INTO mission_progress (id, user_goal_id, mission_id, user_id, is_completed, completed_at, total_tasks, completed_tasks, completion_percentage, created_at, updated_at) VALUES ($1,$2,$3,$4,$5, CASE WHEN $5 THEN now() ELSE NULL END, $6, $7, $8, now(), now())`
	_, err := q.DB.Exec(query, id, userGoalID, missionID, userID, isCompleted, total, completed, perc)
	if err != nil {
		return errors.New("unable to insert mission progress, DB error")
	}
	return nil
}
