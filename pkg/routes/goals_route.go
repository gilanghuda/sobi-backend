package routes

import (
	"github.com/gilanghuda/sobi-backend/app/controllers"
	"github.com/gilanghuda/sobi-backend/pkg/middleware"
	"github.com/gofiber/fiber/v2"
)

func RegisterGoalsRoutes(app *fiber.App) {
	goal := app.Group("/goals", middleware.JWTProtected())
	goal.Post("/create", controllers.CreateUserGoal)
	goal.Get("/today-mission", controllers.GetTodayMissions)
	// admin or internal endpoints to manage missions/tasks
	goal.Post("/missions", controllers.CreateMission)
	goal.Post("/missions/tasks", controllers.CreateTask)
	goal.Post("/tasks/complete", controllers.CompleteTask)
	goal.Post("/summaries", controllers.CreateGoalSummary)
	goal.Get("/summaries", controllers.GetGoalSummaries)
}
