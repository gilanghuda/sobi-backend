package routes

import (
	"github.com/gilanghuda/sobi-backend/app/controllers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

func RegisterChatRoutes(app *fiber.App) {
	chat := app.Group("/chat")
	chat.Post("/rooms", controllers.CreateRoom)
	chat.Get("/rooms", controllers.GetRoomsByUser)
	chat.Post("/messages", controllers.PostMessage)
	chat.Get("/messages", controllers.GetMessagesByRoom)
	chat.Post("/find-match", controllers.FindMatch)

	// Use fasthttp adaptor to wire the net/http MatchmakingHandler to Fiber
	adapter := fasthttpadaptor.NewFastHTTPHandlerFunc(controllers.MatchmakingHandler)
	chat.Post("/matchmaking", func(c *fiber.Ctx) error {
		adapter(c.Context())
		return nil
	})

	chat.Get("/recent", controllers.GetRecentChats)
	chat.Get("/recent/target", controllers.GetRecentChatsAsTarget)
	chat.Get("/active", controllers.GetActiveRoom)
	chat.Post("/bot-message", controllers.ChatWithGemini)

	chat.Get("/ws", websocket.New(func(c *websocket.Conn) {
		controllers.WsHandlerFiber(c)
	}))

	// debug routes
	dbg := chat.Group("/debug")
	dbg.Get("/state", controllers.DebugState)
	dbg.Post("/notify", controllers.DebugNotify)
}
