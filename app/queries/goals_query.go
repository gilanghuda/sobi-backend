package queries

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/google/uuid"
)

type GoalsQueries struct {
	DB *sql.DB
}

func (q *GoalsQueries) CreateUserGoal(g *models.UserGoal) error {
	query := `INSERT INTO user_goals (id, user_id, goal_category, start_date, target_end_date) VALUES ($1, $2, $3, $4, $5)`
	_, err := q.DB.Exec(query, g.ID, g.UserID, g.GoalCategory, g.StartDate, g.TargetEndDate)
	if err != nil {
		println(err.Error())
		return errors.New("unable to create user goal, DB error")
	}
	return nil
}

func (q *GoalsQueries) GetUserGoalsByUser(userID uuid.UUID) ([]models.UserGoal, error) {
	var goals []models.UserGoal
	query := `SELECT id, user_id, goal_category, start_date, target_end_date FROM user_goals WHERE user_id = $1`
	rows, err := q.DB.Query(query, userID)
	if err != nil {
		println(err.Error())
		return goals, errors.New("unable to query user goals")
	}
	defer rows.Close()
	for rows.Next() {
		var g models.UserGoal
		if err := rows.Scan(&g.ID, &g.UserID, &g.GoalCategory, &g.StartDate, &g.TargetEndDate); err != nil {
			return goals, err
		}
		goals = append(goals, g)
	}
	return goals, nil
}

func (q *GoalsQueries) GetUserGoalByID(id uuid.UUID) (models.UserGoal, error) {
	g := models.UserGoal{}
	query := `SELECT id, user_id, goal_category, start_date, target_end_date FROM user_goals WHERE id = $1`
	err := q.DB.QueryRow(query, id).Scan(&g.ID, &g.UserID, &g.GoalCategory, &g.StartDate, &g.TargetEndDate)
	if err != nil {
		if err == sql.ErrNoRows {
			return g, errors.New("user goal not found")
		}
		println(err.Error())
		return g, errors.New("unable to get user goal")
	}
	return g, nil
}

func (q *GoalsQueries) CountMissionsForDayRange(days int) (int, error) {
	var cnt int
	query := `SELECT count(*) FROM missions WHERE day_number <= $1`
	if err := q.DB.QueryRow(query, days).Scan(&cnt); err != nil {
		return 0, errors.New("unable to count missions")
	}
	return cnt, nil
}

func (q *GoalsQueries) CountTasksForMissions(category string, days int) (int, error) {
	var cnt int
	query := `SELECT count(t.id) FROM tasks t JOIN missions m ON t.mission_id = m.id WHERE m.day_number <= $1 AND m.category = $2`
	if err := q.DB.QueryRow(query, days, category).Scan(&cnt); err != nil {
		return 0, errors.New("unable to count tasks")
	}
	return cnt, nil
}

func (q *GoalsQueries) CountMissionProgressCompleted(userGoalID uuid.UUID, days int, userID uuid.UUID) (int, error) {
	var cnt int
	query := `SELECT count(mp.id) FROM mission_progress mp JOIN missions m ON mp.mission_id = m.id WHERE mp.user_goal_id = $1 AND m.day_number <= $2 AND mp.user_id = $3 AND mp.is_completed = true`
	if err := q.DB.QueryRow(query, userGoalID, days, userID).Scan(&cnt); err != nil {
		return 0, errors.New("unable to count completed missions")
	}
	return cnt, nil
}

func (q *GoalsQueries) CountTaskProgressCompleted(userGoalID uuid.UUID, days int, userID uuid.UUID) (int, error) {
	var cnt int
	query := `SELECT count(tp.id) FROM task_progress tp JOIN tasks t ON tp.task_id = t.id JOIN missions m ON t.mission_id = m.id WHERE tp.user_goal_id = $1 AND m.day_number <= $2 AND tp.user_id = $3 AND tp.is_completed = true`
	if err := q.DB.QueryRow(query, userGoalID, days, userID).Scan(&cnt); err != nil {
		return 0, errors.New("unable to count completed tasks")
	}
	return cnt, nil
}

func (q *GoalsQueries) InsertGoalSummary(s *models.GoalSummary) error {
	js := json.RawMessage(s.SelfChanges)
	query := `INSERT INTO goal_summaries (id, user_goal_id, user_id, goal_category, total_days, days_completed, total_missions, missions_completed, total_tasks, tasks_completed, completion_percentage, reflection, self_changes, start_date, end_date, created_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`
	_, err := q.DB.Exec(query, s.ID, s.UserGoalID, s.UserID, s.GoalCategory, s.TotalDays, s.DaysCompleted, s.TotalMissions, s.MissionsCompleted, s.TotalTasks, s.TasksCompleted, s.CompletionPercentage, s.ReflectionText, js, s.StartDate, s.EndDate, s.CreatedAt)
	if err != nil {
		return errors.New("unable to insert goal summary")
	}
	return nil
}

func (q *GoalsQueries) GetGoalSummariesByUserGoal(userGoalID, userID uuid.UUID) ([]models.GoalSummary, error) {
	var res []models.GoalSummary
	query := `SELECT id, user_goal_id, user_id, goal_category, total_days, days_completed, total_missions, missions_completed, total_tasks, tasks_completed, completion_percentage, reflection, self_changes, start_date, end_date, created_at FROM goal_summaries WHERE user_goal_id = $1 AND user_id = $2 ORDER BY created_at DESC`
	rows, err := q.DB.Query(query, userGoalID, userID)
	if err != nil {
		return res, errors.New("unable to query goal summaries")
	}
	defer rows.Close()
	for rows.Next() {
		var s models.GoalSummary
		var selfChangesBytes []byte
		if err := rows.Scan(&s.ID, &s.UserGoalID, &s.UserID, &s.GoalCategory, &s.TotalDays, &s.DaysCompleted, &s.TotalMissions, &s.MissionsCompleted, &s.TotalTasks, &s.TasksCompleted, &s.CompletionPercentage, &s.ReflectionText, &selfChangesBytes, &s.StartDate, &s.EndDate, &s.CreatedAt); err != nil {
			return res, err
		}
		s.SelfChanges = json.RawMessage(selfChangesBytes)
		res = append(res, s)
	}
	return res, nil
}

func (q *GoalsQueries) GetGoalSummariesByUser(userID uuid.UUID) ([]models.GoalSummary, error) {
	var res []models.GoalSummary
	query := `SELECT id, user_goal_id, user_id, goal_category, total_days, days_completed, total_missions, missions_completed, total_tasks, tasks_completed, completion_percentage, reflection, self_changes, start_date, end_date, created_at FROM goal_summaries WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := q.DB.Query(query, userID)
	if err != nil {
		return res, errors.New("unable to query goal summaries")
	}
	defer rows.Close()
	for rows.Next() {
		var s models.GoalSummary
		var selfChangesBytes []byte
		if err := rows.Scan(&s.ID, &s.UserGoalID, &s.UserID, &s.GoalCategory, &s.TotalDays, &s.DaysCompleted, &s.TotalMissions, &s.MissionsCompleted, &s.TotalTasks, &s.TasksCompleted, &s.CompletionPercentage, &s.ReflectionText, &selfChangesBytes, &s.StartDate, &s.EndDate, &s.CreatedAt); err != nil {
			return res, err
		}
		s.SelfChanges = json.RawMessage(selfChangesBytes)
		res = append(res, s)
	}
	return res, nil
}

func (q *GoalsQueries) GetMaxMissionDayForCategory(category string) (int, error) {
	var maxDay sql.NullInt64
	query := `SELECT max(day_number) FROM missions WHERE category = $1`
	if err := q.DB.QueryRow(query, category).Scan(&maxDay); err != nil {
		return 0, errors.New("unable to query max mission day")
	}
	if !maxDay.Valid {
		return 0, nil
	}
	return int(maxDay.Int64), nil
}

func (q *GoalsQueries) CountMissionsForCategory(category string) (int, error) {
	var cnt int
	query := `SELECT count(*) FROM missions WHERE category = $1`
	if err := q.DB.QueryRow(query, category).Scan(&cnt); err != nil {
		return 0, errors.New("unable to count missions for category")
	}
	return cnt, nil
}

func (q *GoalsQueries) CountTasksForCategory(category string) (int, error) {
	var cnt int
	query := `SELECT count(t.id) FROM tasks t JOIN missions m ON t.mission_id = m.id WHERE m.category = $1`
	if err := q.DB.QueryRow(query, category).Scan(&cnt); err != nil {
		return 0, errors.New("unable to count tasks for category")
	}
	return cnt, nil
}

func (q *GoalsQueries) BuildGoalSummary(ug models.UserGoal, userID uuid.UUID) (models.GoalSummary, error) {
	var s models.GoalSummary

	totalDays := int(ug.TargetEndDate.Sub(ug.StartDate).Hours()/24) + 1
	if totalDays < 0 {
		totalDays = 0
	}

	dayIndex := totalDays
	if dayIndex < 1 {
		dayIndex = 1
	}

	totalMissions := totalDays
	totalTasks, err := q.CountTasksForMissions(ug.GoalCategory, totalDays)
	if err != nil {
		return s, err
	}

	missionsCompleted, err := q.CountMissionProgressCompleted(ug.ID, dayIndex, userID)
	if err != nil {
		return s, err
	}
	tasksCompleted, err := q.CountTaskProgressCompleted(ug.ID, dayIndex, userID)
	if err != nil {
		return s, err
	}

	perc := 0.0
	if totalTasks > 0 {
		perc = (float64(tasksCompleted) / float64(totalTasks)) * 100
		perc = mathRound(perc, 2)
	}

	emptySelf := json.RawMessage([]byte("[]"))
	s = models.GoalSummary{
		ID:                   uuid.New(),
		UserGoalID:           ug.ID,
		UserID:               userID,
		GoalCategory:         ug.GoalCategory,
		TotalDays:            totalDays,
		DaysCompleted:        dayIndex,
		TotalMissions:        totalMissions,
		MissionsCompleted:    missionsCompleted,
		TotalTasks:           totalTasks,
		TasksCompleted:       tasksCompleted,
		CompletionPercentage: perc,
		ReflectionText:       "",
		SelfChanges:          emptySelf,
		StartDate:            ug.StartDate,
		EndDate:              ug.TargetEndDate,
		CreatedAt:            time.Now(),
	}
	return s, nil
}

func mathRound(val float64, places int) float64 {
	factor := 1.0
	for i := 0; i < places; i++ {
		factor *= 10
	}
	return float64(int(val*factor+0.5)) / factor
}
