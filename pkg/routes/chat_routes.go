package routes

import (
	"github.com/gilanghuda/sobi-backend/app/controllers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func RegisterChatRoutes(app *fiber.App) {
	chat := app.Group("/chat")
	chat.Post("/rooms", controllers.CreateRoom)
	chat.Get("/rooms", controllers.GetRoomsByUser)
	chat.Post("/messages", controllers.PostMessage)
	chat.Get("/messages", controllers.GetMessagesByRoom)
	chat.Post("/match", controllers.MatchHandler)
	chat.Get("/recent", controllers.GetRecentChats)
	chat.Get("/recent/target", controllers.GetRecentChatsAsTarget)
	chat.Get("/active", controllers.GetActiveRoom)
	chat.Post("/bot-message", controllers.ChatWithGemini)

	chat.Get("/ws", websocket.New(func(c *websocket.Conn) {
		controllers.WsHandler(c)
	}))
}
