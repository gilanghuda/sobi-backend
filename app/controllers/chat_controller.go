package controllers

import (
	"encoding/json"
	"time"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gilanghuda/sobi-backend/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

type client struct {
	conn *websocket.Conn
	uid  uuid.UUID
}

var hubs = make(map[string]map[*client]bool)

func serveWs(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		return c.Next()
	}
	return c.Status(fiber.StatusUpgradeRequired).SendString("upgrade required")
}

func WsHandler(c *websocket.Conn) {
	roomIDStr := c.Query("room_id")
	if roomIDStr == "" {
		c.WriteMessage(websocket.TextMessage, []byte("room_id required"))
		c.Close()
		return
	}
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		c.WriteMessage(websocket.TextMessage, []byte("invalid room_id"))
		c.Close()
		return
	}
	token := c.Query("token")
	var userID uuid.UUID
	if token != "" {
		head := "Bearer " + token
		userID, _ = utils.ExtractUserIDFromHeader(head)
	}

	cl := &client{conn: c, uid: userID}
	if _, ok := hubs[roomID.String()]; !ok {
		hubs[roomID.String()] = make(map[*client]bool)
	}
	hubs[roomID.String()][cl] = true

	defer func() {
		delete(hubs[roomID.String()], cl)
		c.Close()
	}()

	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		for cc := range hubs[roomID.String()] {
			if cc == cl {
				continue
			}
			cc.conn.WriteMessage(mt, msg)
		}
		_ = msg
	}
}

func CreateRoom(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	req := &models.CreateRoomRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}

	r := &models.Room{ID: uuid.New(), OwnerID: userID, Category: req.Category, Visible: req.Visible, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	q := queries.ChatQueries{DB: database.DB}
	if err := q.CreateRoom(r); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create room"})
	}
	return c.Status(fiber.StatusCreated).JSON(r)
}

func GetRoomsByUser(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	q := queries.ChatQueries{DB: database.DB}
	rooms, err := q.GetRoomsByUser(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get rooms"})
	}
	return c.Status(fiber.StatusOK).JSON(rooms)
}

func PostMessage(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	p := &models.CreateMessageRequest{}
	if err := c.BodyParser(p); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	roomID, err := uuid.Parse(p.RoomID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room_id"})
	}
	vis := true
	if p.Visible != nil {
		vis = *p.Visible
	}
	m := &models.Message{ID: uuid.New(), RoomID: roomID, UserID: userID, Text: p.Text, Visible: vis, CreatedAt: time.Now()}
	q := queries.ChatQueries{DB: database.DB}
	if err := q.CreateMessage(m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create message"})
	}

	if clients, ok := hubs[roomID.String()]; ok {
		b, _ := json.Marshal(m)
		for cl := range clients {
			cl.conn.WriteMessage(websocket.TextMessage, b)
		}
	}
	return c.Status(fiber.StatusCreated).JSON(m)
}

func GetMessagesByRoom(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	_, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	roomIDStr := c.Query("room_id")
	if roomIDStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "room_id required"})
	}
	roomID, err := uuid.Parse(roomIDStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid room_id"})
	}
	limit := 100
	q := queries.ChatQueries{DB: database.DB}
	msgs, err := q.GetMessagesByRoom(roomID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get messages"})
	}
	return c.Status(fiber.StatusOK).JSON(msgs)
}
