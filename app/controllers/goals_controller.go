package controllers

import (
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gilanghuda/sobi-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func CreateUserGoal(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	req := &models.CreateUserGoalRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid start_date format, use YYYY-MM-DD"})
	}
	targetDate, err := time.Parse("2006-01-02", req.TargetEndDate)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid target_end_date format, use YYYY-MM-DD"})
	}
	if !targetDate.After(startDate) && !targetDate.Equal(startDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "target_end_date must be same or after start_date"})
	}

	gq := queries.GoalsQueries{DB: database.DB}
	existingGoals, err := gq.GetUserGoalsByUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to query existing user goals"})
	}
	for _, eg := range existingGoals {
		egEnd := time.Date(eg.TargetEndDate.Year(), eg.TargetEndDate.Month(), eg.TargetEndDate.Day(), 0, 0, 0, 0, eg.TargetEndDate.Location())
		newStartOnly := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		if !egEnd.Before(newStartOnly) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "you already have an active goal in the requested period"})
		}
	}

	g := &models.UserGoal{
		ID:            uuid.New(),
		UserID:        userID,
		GoalCategory:  req.GoalCategory,
		Status:        "active",
		CurrentDay:    1,
		StartDate:     startDate,
		TargetEndDate: targetDate,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := gq.CreateUserGoal(g); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create user goal"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "User goal created", "goal": g})
}

func GetMissions(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	ugq := queries.GoalsQueries{DB: database.DB}
	userGoals, err := ugq.GetUserGoalsByUser(userID)
	if err != nil {
		println(err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get user goals"})
	}

	mQ := queries.MissionsQueries{DB: database.DB}

	dateStr := strings.TrimSpace(c.Query("date"))
	if dateStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "date query param required (YYYY-MM-DD)"})
	}
	inputDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid date format, use YYYY-MM-DD"})
	}

	inputOnly := time.Date(inputDate.Year(), inputDate.Month(), inputDate.Day(), 0, 0, 0, 0, inputDate.Location())

	var result []map[string]interface{}

	for _, ug := range userGoals {
		computedStatus := "active"
		if inputOnly.After(ug.TargetEndDate) {
			summaries, _ := ugq.GetGoalSummariesByUserGoal(ug.ID, userID)
			if len(summaries) == 0 {
				computedStatus = "completed"
			} else {
				computedStatus = "inactive"
			}
		}

		if computedStatus == "inactive" {
			continue
		}
		ug.Status = computedStatus
		startDate := ug.StartDate

		startY, startM, startD := startDate.Date()

		startOnly := time.Date(startY, startM, startD, 0, 0, 0, 0, startDate.Location())

		days := int(inputOnly.Sub(startOnly).Hours()/24) + 1
		if days < 1 {
			days = 1
		}

		missions, err := mQ.GetMissionsByDayAndCategory(days, ug.GoalCategory)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get missions"})
		}

		var mwts []models.MissionWithTasks
		for _, m := range missions {
			tasks, err := mQ.GetTasksByMission(m.ID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get tasks"})
			}

			var tps []models.TaskWithProgress
			for _, t := range tasks {
				p, err := mQ.GetTaskProgress(ug.ID, t.ID, userID)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get task progress"})
				}
				tps = append(tps, models.TaskWithProgress{Task: t, Progress: p})
			}

			mp, err := mQ.GetMissionProgress(ug.ID, m.ID, userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get mission progress"})
			}

			mwts = append(mwts, models.MissionWithTasks{Mission: m, Tasks: tps, Progress: mp})
		}

		var prevDays []map[string]interface{}
		for d := 1; d < days; d++ {
			dayMissions, err := mQ.GetMissionsByDayAndCategory(d, ug.GoalCategory)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get missions"})
			}
			dayComplete := true
			for _, dm := range dayMissions {
				mp, err := mQ.GetMissionProgress(ug.ID, dm.ID, userID)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get mission progress"})
				}
				if mp.ID == uuid.Nil || !mp.IsCompleted {
					dayComplete = false
					break
				}
			}
			prevDays = append(prevDays, map[string]interface{}{"day_index": d, "is_completed": dayComplete})
		}

		result = append(result, map[string]interface{}{
			"user_goal":     ug,
			"day_index":     days,
			"missions":      mwts,
			"previous_days": prevDays,
		})
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

func CreateMission(c *fiber.Ctx) error {
	payload := struct {
		DayNumber int    `json:"day_number"`
		Focus     string `json:"focus"`
		Category  string `json:"category"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	m := &models.Mission{ID: uuid.New(), DayNumber: payload.DayNumber, Focus: payload.Focus, Category: payload.Category}
	mq := queries.MissionsQueries{DB: database.DB}
	if err := mq.CreateMission(m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create mission"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Mission created", "mission": m})
}

func CreateTask(c *fiber.Ctx) error {
	payload := struct {
		MissionID string `json:"mission_id"`
		Text      string `json:"text"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}
	mid, err := uuid.Parse(payload.MissionID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid mission_id"})
	}
	t := &models.Task{ID: uuid.New(), MissionID: mid, Text: payload.Text}
	mq := queries.MissionsQueries{DB: database.DB}
	if err := mq.CreateTask(t); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create task"})
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Task created", "task": t})
}

func CompleteTask(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	payload := struct {
		UserGoalID string `json:"user_goal_id"`
		TaskID     string `json:"task_id"`
		Completed  bool   `json:"completed"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ugid, err := uuid.Parse(payload.UserGoalID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user_goal_id"})
	}
	tid, err := uuid.Parse(payload.TaskID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid task_id"})
	}

	mq := queries.MissionsQueries{DB: database.DB}

	existing, err := mq.GetTaskProgress(ugid, tid, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get task progress"})
	}
	if existing.ID != uuid.Nil {
		if _, err := mq.UpdateTaskProgress(ugid, tid, userID, payload.Completed); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update task progress"})
		}
	} else {
		if err := mq.InsertTaskProgress(uuid.New(), ugid, tid, userID, payload.Completed); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert task progress"})
		}
	}

	var missionID uuid.UUID
	if err := database.DB.QueryRow(`SELECT mission_id FROM tasks WHERE id = $1`, tid).Scan(&missionID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find mission for task"})
	}

	total, err := mq.CountTotalTasks(missionID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count total tasks"})
	}
	completed, err := mq.CountCompletedTasks(ugid, missionID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count completed tasks"})
	}
	perc := 0.0
	if total > 0 {
		perc = (float64(completed) / float64(total)) * 100
	}
	isCompleted := false
	if total > 0 && completed == total {
		isCompleted = true
	}

	if rows, err := mq.UpdateMissionProgress(ugid, missionID, userID, total, completed, perc, isCompleted); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update mission progress"})
	} else if rows == 0 {
		if err := mq.InsertMissionProgress(uuid.New(), ugid, missionID, userID, total, completed, perc, isCompleted); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert mission progress"})
		}
	}

	p2, err := mq.GetTaskProgress(ugid, tid, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get task progress"})
	}
	mp, err := mq.GetMissionProgress(ugid, missionID, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get mission progress"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"task_progress": p2, "mission_progress": mp})
}

func CreateGoalSummary(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	payload := &models.CreateGoalSummaryRequest{}
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	ugid, err := uuid.Parse(payload.UserGoalID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user_goal_id"})
	}

	gq := queries.GoalsQueries{DB: database.DB}
	ug, err := gq.GetUserGoalByID(ugid)
	if err != nil {
		println(err.Error())
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User goal not found"})
	}

	totalDays := int(ug.TargetEndDate.Sub(ug.StartDate).Hours()/24) + 1
	if totalDays < 0 {
		totalDays = 0
	}

	dayIndex := totalDays
	if dayIndex < 1 {
		dayIndex = 1
	}

	totalMissions := totalDays
	totalTasks, err := gq.CountTasksForMissions(ug.GoalCategory, totalDays)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count tasks"})
	}

	missionsCompleted, err := gq.CountMissionProgressCompleted(ugid, dayIndex, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count completed missions"})
	}
	tasksCompleted, err := gq.CountTaskProgressCompleted(ugid, dayIndex, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to count completed tasks"})
	}

	perc := 0.0
	if totalTasks > 0 {
		perc = (float64(tasksCompleted) / float64(totalTasks)) * 100
		perc = math.Round(perc*100) / 100
	}

	js, _ := json.Marshal(payload.SelfChanges)
	s := &models.GoalSummary{
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
		ReflectionText:       strings.Join(payload.Reflection, "\n"),
		SelfChanges:          json.RawMessage(js),
		StartDate:            ug.StartDate,
		EndDate:              ug.TargetEndDate,
		CreatedAt:            time.Now(),
	}

	if err := gq.InsertGoalSummary(s); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to insert goal summary"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"message": "Goal summary created", "summary": s})
}

func GetGoalSummaries(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	userGoalIDStr := strings.TrimSpace(c.Query("user_goal_id"))

	gq := queries.GoalsQueries{DB: database.DB}
	var summaries []models.GoalSummary
	if userGoalIDStr == "" {
		summaries, err = gq.GetGoalSummariesByUser(userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get goal summaries"})
		}
		if len(summaries) == 0 {
			ugq := queries.GoalsQueries{DB: database.DB}
			ugs, _ := ugq.GetUserGoalsByUser(userID)
			for _, ug := range ugs {
				summary, err := gq.BuildGoalSummary(ug, userID)
				if err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build goal summary"})
				}
				summaries = append(summaries, summary)
			}
		}
	} else {
		ugid, err := uuid.Parse(userGoalIDStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid user_goal_id"})
		}

		summaries, err = gq.GetGoalSummariesByUserGoal(ugid, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get goal summaries"})
		}
		if len(summaries) == 0 {
			ug, err := gq.GetUserGoalByID(ugid)
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User goal not found"})
			}
			summary, err := gq.BuildGoalSummary(ug, userID)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to build goal summary"})
			}
			summaries = append(summaries, summary)
		}
	}

	loc, locErr := time.LoadLocation("Asia/Jakarta")
	if locErr != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	formatTime := func(t time.Time) string {
		if t.IsZero() {
			return ""
		}
		return t.In(loc).Format(time.RFC3339)
	}

	var out []map[string]interface{}
	for _, s := range summaries {
		var selfChanges []string
		_ = json.Unmarshal(s.SelfChanges, &selfChanges)
		out = append(out, map[string]interface{}{
			"id":                    s.ID,
			"user_goal_id":          s.UserGoalID,
			"user_id":               s.UserID,
			"goal_category":         s.GoalCategory,
			"total_days":            s.TotalDays,
			"days_completed":        s.DaysCompleted,
			"total_missions":        s.TotalMissions,
			"missions_completed":    s.MissionsCompleted,
			"total_tasks":           s.TotalTasks,
			"tasks_completed":       s.TasksCompleted,
			"completion_percentage": s.CompletionPercentage,
			"reflection":            s.ReflectionText,
			"self_changes":          selfChanges,
			"start_date":            formatTime(s.StartDate),
			"end_date":              formatTime(s.EndDate),
			"created_at":            formatTime(s.CreatedAt),
		})
	}

	return c.Status(fiber.StatusOK).JSON(out)
}
