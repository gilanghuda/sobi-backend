package controllers

import (
	"encoding/json"
	"sync"
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

var hub = make(map[*client]bool)

var pendengarQueue = make(map[string][]uuid.UUID)
var penceritaQueue = make(map[string][]uuid.UUID)

var messageChan = make(chan *models.Message, 100)
var clientsByUser = make(map[uuid.UUID]map[*client]bool)
var clientsMu sync.RWMutex

func init() {
	go func() {
		for m := range messageChan {
			q := queries.ChatQueries{DB: database.DB}
			r, err := q.GetRoomByID(m.RoomID)
			if err != nil {
				continue
			}
			recipients := []uuid.UUID{r.OwnerID}
			if r.TargetID != nil {
				recipients = append(recipients, *r.TargetID)
			}

			b, _ := json.Marshal(m)
			clientsMu.RLock()
			for _, uid := range recipients {
				if conns, ok := clientsByUser[uid]; ok {
					for cc := range conns {
						if cc.conn != nil {
							cc.conn.WriteMessage(websocket.TextMessage, b)
						}
					}
				}
			}
			clientsMu.RUnlock()
		}
	}()
}

func WsHandler(c *websocket.Conn) {
	token := c.Query("token")
	var userID uuid.UUID
	if token != "" {
		head := "Bearer " + token
		userID, _ = utils.ExtractUserIDFromHeader(head)
	}

	cl := &client{conn: c, uid: userID}
	hub[cl] = true

	clientsMu.Lock()
	if cl.uid != uuid.Nil {
		if _, ok := clientsByUser[cl.uid]; !ok {
			clientsByUser[cl.uid] = make(map[*client]bool)
		}
		clientsByUser[cl.uid][cl] = true
	}
	clientsMu.Unlock()

	defer func() {
		delete(hub, cl)
		clientsMu.Lock()
		if cl.uid != uuid.Nil {
			if conns, ok := clientsByUser[cl.uid]; ok {
				delete(conns, cl)
				if len(conns) == 0 {
					delete(clientsByUser, cl.uid)
				}
			}
		}
		clientsMu.Unlock()
		c.Close()
	}()

	for {
		mt, msg, err := c.ReadMessage()
		if err != nil {
			break
		}
		for cc := range hub {
			if cc == cl {
				continue
			}
			cc.conn.WriteMessage(mt, msg)
		}
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

	var targetPtr *uuid.UUID
	if req.TargetID != "" {
		if tid, err := uuid.Parse(req.TargetID); err == nil {
			targetPtr = &tid
		}
	}

	r := &models.Room{ID: uuid.New(), OwnerID: userID, TargetID: targetPtr, Category: req.Category, Visible: req.Visible, CreatedAt: time.Now(), UpdatedAt: time.Now()}
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
	m := &models.Message{
		ID:        uuid.New(),
		RoomID:    roomID,
		UserID:    userID,
		Text:      p.Text,
		Visible:   vis,
		CreatedAt: time.Now()}
	q := queries.ChatQueries{DB: database.DB}
	if err := q.CreateMessage(m); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create message"})
	}

	go func(msg *models.Message) {
		messageChan <- msg
	}(m)

	_, _ = json.Marshal(m)
	return c.Status(fiber.StatusCreated).JSON(m)
}

func GetMessagesByRoom(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
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

	type Message struct {
		ID        uuid.UUID `json:"id"`
		RoomID    uuid.UUID `json:"room_id"`
		UserID    uuid.UUID `json:"user_id"`
		Text      string    `json:"text"`
		Visible   bool      `json:"visible"`
		CreatedAt time.Time `json:"created_at"`
		IsMe      bool      `json:"is_me"`
	}

	out := make([]Message, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, Message{
			ID:        m.ID,
			RoomID:    m.RoomID,
			UserID:    m.UserID,
			Text:      m.Text,
			Visible:   m.Visible,
			CreatedAt: m.CreatedAt,
			IsMe:      m.UserID == userID,
		})
	}

	return c.Status(fiber.StatusOK).JSON(out)
}

func GetRecentChats(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	limit := 50
	q := queries.ChatQueries{DB: database.DB}
	recent, err := q.GetRecentChats(userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get recent chats"})
	}
	return c.Status(fiber.StatusOK).JSON(recent)
}

func GetRecentChatsAsTarget(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}
	limit := 50
	q := queries.ChatQueries{DB: database.DB}
	recent, err := q.GetRecentChatsAsTarget(userID, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get recent chats"})
	}
	return c.Status(fiber.StatusOK).JSON(recent)
}

func GetActiveRoom(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	payload := struct {
		Time string `json:"time"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if payload.Time == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "time required"})
	}

	var t time.Time
	var parseErr error
	layouts := []string{"2006-01-02 15:04", "2006-01-02 15:04:05", "2006-01-02"}
	for _, lay := range layouts {
		t, parseErr = time.Parse(lay, payload.Time)
		if parseErr == nil {
			break
		}
	}
	if parseErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid time format. use YYYY-MM-DD or YYYY-MM-DD HH:MM[:SS]"})
	}

	endTime := t
	startTime := t.Add(-30 * time.Minute)

	q := queries.ChatQueries{DB: database.DB}
	r, err := q.GetActiveRoom(userID, startTime, endTime)
	if err != nil {
		if err.Error() == "no active room" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "no active room"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to get active room"})
	}
	return c.Status(fiber.StatusOK).JSON(r)
}

func MatchHandler(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	payload := &models.MatchRequest{}
	if err := c.BodyParser(payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if payload.Role != "pendengar" && payload.Role != "pencerita" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "role must be pendengar or pencerita"})
	}
	cat := payload.Category
	if cat == "" {
		cat = "default"
	}

	if payload.Role == "pendengar" {
		if q, ok := penceritaQueue[cat]; ok && len(q) > 0 {
			otherID := q[0]
			penceritaQueue[cat] = q[1:]
			r := &models.Room{ID: uuid.New(), OwnerID: userID, TargetID: &otherID, Category: cat, Visible: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			qq := queries.ChatQueries{DB: database.DB}
			if err := qq.CreateRoom(r); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create room"})
			}
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"room_id": r.ID, "matched_with": otherID})
		}

		pendengarQueue[cat] = append(pendengarQueue[cat], userID)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "queued"})
	}

	if q, ok := pendengarQueue[cat]; ok && len(q) > 0 {
		otherID := q[0]
		pendengarQueue[cat] = q[1:]
		r := &models.Room{ID: uuid.New(), OwnerID: otherID, TargetID: &userID, Category: cat, Visible: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		qq := queries.ChatQueries{DB: database.DB}
		if err := qq.CreateRoom(r); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create room"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"room_id": r.ID, "matched_with": otherID})
	}

	penceritaQueue[cat] = append(penceritaQueue[cat], userID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "queued"})
}

func ChatWithGemini(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	// allow both authenticated and anonymous usage
	_, _ = utils.ExtractUserIDFromHeader(authHeader)

	payload := struct {
		Prompt string `json:"prompt"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if payload.Prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "prompt required"})
	}

	// simple one-shot request to Gemini (no history persistence)
	reply, err := utils.QueryGemini(payload.Prompt)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "gemini error: " + err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"reply": reply})
}
