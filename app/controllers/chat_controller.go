package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
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

// simple in-memory matchmaking queues per category
var pendengarQueue = make(map[string][]uuid.UUID) // category -> []userID
var penceritaQueue = make(map[string][]uuid.UUID) // category -> []userID

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

// MatchHandler pairs users based on role and category
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

	// try to find opposite role
	if payload.Role == "pendengar" {
		// check pencerita queue
		if q, ok := penceritaQueue[cat]; ok && len(q) > 0 {
			otherID := q[0]
			penceritaQueue[cat] = q[1:]
			// create room
			r := &models.Room{ID: uuid.New(), OwnerID: otherID, Category: cat, Visible: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			qq := queries.ChatQueries{DB: database.DB}
			if err := qq.CreateRoom(r); err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create room"})
			}
			return c.Status(fiber.StatusOK).JSON(fiber.Map{"room_id": r.ID, "matched_with": otherID})
		}
		// enqueue
		pendengarQueue[cat] = append(pendengarQueue[cat], userID)
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "queued"})
	}

	// role pencerita
	if q, ok := pendengarQueue[cat]; ok && len(q) > 0 {
		otherID := q[0]
		pendengarQueue[cat] = q[1:]
		r := &models.Room{ID: uuid.New(), OwnerID: otherID, Category: cat, Visible: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}
		qq := queries.ChatQueries{DB: database.DB}
		if err := qq.CreateRoom(r); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to create room"})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"room_id": r.ID, "matched_with": otherID})
	}
	// enqueue
	penceritaQueue[cat] = append(penceritaQueue[cat], userID)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"message": "queued"})
}

// ChatWithGemini proxies a user prompt to Gemini API and returns the assistant response.
func ChatWithGemini(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	userID, err := utils.ExtractUserIDFromHeader(authHeader)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
	}

	payload := struct {
		Prompt string `json:"prompt"`
		Model  string `json:"model,omitempty"`
		RoomID string `json:"room_id,omitempty"`
	}{}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	if payload.Prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "prompt is required"})
	}

	var qq queries.ChatQueries
	if payload.RoomID != "" {
		// parse room id and collect recent messages as context
		if rid, err := uuid.Parse(payload.RoomID); err == nil {
			// persist user's message
			um := &models.Message{ID: uuid.New(), RoomID: rid, UserID: userID, Text: payload.Prompt, Visible: true, CreatedAt: time.Now()}
			_ = qq.CreateMessage(um)
			// broadcast user's message
			if clients, ok := hubs[rid.String()]; ok {
				b, _ := json.Marshal(um)
				for cl := range clients {
					cl.conn.WriteMessage(websocket.TextMessage, b)
				}
			}

			// build context array
			var contextTexts []string
			msgs, _ := qq.GetMessagesByRoom(rid, 50)
			for _, m := range msgs {
				contextTexts = append(contextTexts, m.Text)
			}

			// include last messages in the request body
			reqBody := map[string]interface{}{"prompt": payload.Prompt, "context": contextTexts}
			if payload.Model != "" {
				reqBody["model"] = payload.Model
			}

			// call gemini
			geminiURL := os.Getenv("GEMINI_API_URL")
			if geminiURL == "" {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "GEMINI_API_URL not configured"})
			}
			key := os.Getenv("GEMINI_API_KEY")
			if key == "" {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "GEMINI_API_KEY not configured"})
			}

			b, _ := json.Marshal(reqBody)
			req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, geminiURL, bytes.NewReader(b))
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to prepare request"})
			}
			req.Header.Set("Authorization", "Bearer "+key)
			req.Header.Set("Content-Type", "application/json")

			cli := &http.Client{Timeout: 20 * time.Second}
			resp, err := cli.Do(req)
			if err != nil {
				return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to call gemini"})
			}
			defer resp.Body.Close()
			respBytes, _ := io.ReadAll(resp.Body)

			var gemResp interface{}
			_ = json.Unmarshal(respBytes, &gemResp)

			assistantText := ""
			if m, ok := gemResp.(map[string]interface{}); ok {
				if t, ok2 := m["text"].(string); ok2 {
					assistantText = t
				} else if choices, ok3 := m["choices"].([]interface{}); ok3 && len(choices) > 0 {
					if c0, ok4 := choices[0].(map[string]interface{}); ok4 {
						if t2, ok5 := c0["text"].(string); ok5 {
							assistantText = t2
						}
					}
				}
			}

			// persist assistant message
			am := &models.Message{ID: uuid.New(), RoomID: rid, UserID: uuid.Nil, Text: assistantText, Visible: true, CreatedAt: time.Now()}
			_ = qq.CreateMessage(am)
			// broadcast assistant message
			if clients, ok := hubs[rid.String()]; ok {
				b2, _ := json.Marshal(am)
				for cl := range clients {
					cl.conn.WriteMessage(websocket.TextMessage, b2)
				}
			}

			return c.Status(fiber.StatusOK).JSON(fiber.Map{"assistant": gemResp, "assistant_text": assistantText})
		}
	}

	// if no room context, call gemini directly
	geminiURL := os.Getenv("GEMINI_API_URL")
	if geminiURL == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "GEMINI_API_URL not configured"})
	}
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "GEMINI_API_KEY not configured"})
	}

	reqBody := map[string]interface{}{"prompt": payload.Prompt}
	if payload.Model != "" {
		reqBody["model"] = payload.Model
	}
	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, geminiURL, bytes.NewReader(b))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to prepare request"})
	}
	req.Header.Set("Authorization", "Bearer "+key)
	req.Header.Set("Content-Type", "application/json")

	cli := &http.Client{Timeout: 20 * time.Second}
	resp, err := cli.Do(req)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to call gemini"})
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(resp.Body)

	var gemResp interface{}
	_ = json.Unmarshal(respBytes, &gemResp)

	assistantText := ""
	if m, ok := gemResp.(map[string]interface{}); ok {
		if t, ok2 := m["text"].(string); ok2 {
			assistantText = t
		} else if choices, ok3 := m["choices"].([]interface{}); ok3 && len(choices) > 0 {
			if c0, ok4 := choices[0].(map[string]interface{}); ok4 {
				if t2, ok5 := c0["text"].(string); ok5 {
					assistantText = t2
				}
			}
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"assistant": gemResp, "assistant_text": assistantText})
}
