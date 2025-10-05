package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"

	"github.com/gilanghuda/sobi-backend/pkg/utils"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/gofiber/fiber/v2"
)

// MatchEntry represents a waiting user in the matchmaking queue.
type MatchEntry struct {
	UserID   uuid.UUID
	Role     string
	Category string
	Timer    *time.Timer
}

// MatchMaker manages in-memory queues for categories and roles.
type MatchMaker struct {
	mu     sync.Mutex
	queues map[string]map[string][]*MatchEntry // category -> role -> slice of entries
	ttl    time.Duration
}

// NewMatchMaker creates a new MatchMaker with the provided TTL for entries.
func NewMatchMaker(ttl time.Duration) *MatchMaker {
	m := &MatchMaker{
		queues: make(map[string]map[string][]*MatchEntry),
		ttl:    ttl,
	}
	return m
}

// enqueue adds a user to the queue and starts a timeout to remove them.
func (m *MatchMaker) enqueue(entry *MatchEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.queues[entry.Category]; !ok {
		m.queues[entry.Category] = make(map[string][]*MatchEntry)
	}
	m.queues[entry.Category][entry.Role] = append(m.queues[entry.Category][entry.Role], entry)
	log.Printf("event=enqueue user=%s role=%s category=%s queue_len=%d", entry.UserID, entry.Role, entry.Category, len(m.queues[entry.Category][entry.Role]))

	// start timeout
	entry.Timer = time.AfterFunc(m.ttl, func() {
		m.removeByUser(entry.Category, entry.Role, entry.UserID)
		log.Printf("event=timeout user=%s role=%s category=%s", entry.UserID, entry.Role, entry.Category)
	})
}

// dequeueOpposite tries to find a waiting user with the opposite role in the category.
func (m *MatchMaker) dequeueOpposite(category, role string) *MatchEntry {
	opposite := oppositeRole(role)
	m.mu.Lock()
	defer m.mu.Unlock()
	roleQueue := m.queues[category][opposite]
	if len(roleQueue) == 0 {
		return nil
	}
	// pop first
	entry := roleQueue[0]
	m.queues[category][opposite] = roleQueue[1:]
	if entry.Timer != nil {
		entry.Timer.Stop()
	}
	log.Printf("event=dequeue user=%s role=%s category=%s remaining=%d", entry.UserID, entry.Role, entry.Category, len(m.queues[category][opposite]))
	return entry
}

// removeByUser removes a user from the queue (used for disconnects and timeouts).
func (m *MatchMaker) removeByUser(category, role string, userID uuid.UUID) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// defensive: if category or role map doesn't exist, nothing to do
	rolesMap, ok := m.queues[category]
	if !ok {
		return
	}
	queue := rolesMap[role]
	newQ := make([]*MatchEntry, 0, len(queue))
	removed := false
	for _, e := range queue {
		if e.UserID == userID {
			removed = true
			if e.Timer != nil {
				e.Timer.Stop()
			}
			continue
		}
		newQ = append(newQ, e)
	}
	rolesMap[role] = newQ
	if removed {
		log.Printf("event=removed_from_queue user=%s role=%s category=%s", userID, role, category)
	}
}

func oppositeRole(role string) string {
	if role == "pencerita" {
		return "pendengar"
	}
	return "pencerita"
}

// Global matcher instance (single in-memory instance)
var Matcher = NewMatchMaker(60 * time.Second)

// in-memory room members mapping: room_id -> []userUUID
var roomMembersMu sync.RWMutex
var roomMembers = make(map[string][]uuid.UUID)

// WsHandlerFiber is Fiber-compatible websocket handler. Accepts token query param and extracts user id.
func WsHandlerFiber(c *websocket.Conn) {
	// accept token from query string (frontend may pass JWT here)
	token := c.Query("token")
	var userID uuid.UUID
	if token != "" {
		head := "Bearer " + token
		userID, _ = utils.ExtractUserIDFromHeader(head)
	}

	// register connection (userID may be uuid.Nil if unauthenticated)
	utils.DefaultNotifier.Register(userID, c)
	log.Printf("event=ws_connected user=%s", userID.String())

	// read loop to detect close and to route incoming messages
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			// non-JSON messages are ignored for routing
			continue
		}

		// attach sender id so recipients know who sent it
		if userID != uuid.Nil {
			payload["sender_id"] = userID.String()
		}

		evt, _ := payload["event"].(string)
		if evt == "message" {
			roomID, _ := payload["room_id"].(string)
			if roomID != "" {
				roomMembersMu.RLock()
				members := roomMembers[roomID]
				roomMembersMu.RUnlock()
				for _, member := range members {
					if userID != uuid.Nil && member == userID {
						continue
					}
					_ = utils.DefaultNotifier.Send(member, payload)
				}
			} else {
				// broadcast to all other connected users
				for _, uid := range utils.DefaultNotifier.ActiveUserIDs() {
					if userID != uuid.Nil && uid == userID {
						continue
					}
					_ = utils.DefaultNotifier.Send(uid, payload)
				}
			}
			continue
		}
		// handle other events if needed
	}

	utils.DefaultNotifier.Unregister(userID)

	// Snapshot categories under lock to avoid double-locking when calling removeByUser
	Matcher.mu.Lock()
	cats := make([]string, 0, len(Matcher.queues))
	for cat := range Matcher.queues {
		cats = append(cats, cat)
	}
	Matcher.mu.Unlock()

	for _, cat := range cats {
		Matcher.removeByUser(cat, "pencerita", userID)
		Matcher.removeByUser(cat, "pendengar", userID)
	}

	log.Printf("event=ws_disconnected user=%s", userID.String())
	_ = c.Close()
}

type MatchmakingRequest struct {
	Role     string `json:"role"`
	Category string `json:"category"`
	UserID   string `json:"user_id"`
}

func MatchmakingHandler(w http.ResponseWriter, r *http.Request) {
	var req MatchmakingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	if req.Role != "pencerita" && req.Role != "pendengar" {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	if req.Category == "" {
		http.Error(w, "category required", http.StatusBadRequest)
		return
	}

	opp := Matcher.dequeueOpposite(req.Category, req.Role)
	if opp == nil {
		entry := &MatchEntry{UserID: userID, Role: req.Role, Category: req.Category}
		Matcher.enqueue(entry)
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("waiting"))
		return
	}

	room := &models.Room{
		ID:        uuid.New(),
		OwnerID:   userID,
		TargetID:  &opp.UserID,
		Category:  req.Category,
		Visible:   true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// In-memory placeholder: in production save to DB.
	log.Printf("event=match_success room=%s user1=%s user2=%s category=%s", room.ID, userID, opp.UserID, req.Category)

	// register room members so WS handler can route messages
	roomMembersMu.Lock()
	roomMembers[room.ID.String()] = []uuid.UUID{userID, opp.UserID}
	roomMembersMu.Unlock()

	payload1 := map[string]string{"event": "matched", "room_id": room.ID.String(), "matched_with": opp.UserID.String()}
	payload2 := map[string]string{"event": "matched", "room_id": room.ID.String(), "matched_with": userID.String()}

	if err := utils.DefaultNotifier.Send(userID, payload1); err != nil {
		log.Printf("event=notify_error user=%s err=%v", userID, err)
	}
	if err := utils.DefaultNotifier.Send(opp.UserID, payload2); err != nil {
		log.Printf("event=notify_error user=%s err=%v", opp.UserID, err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "matched", "room_id": room.ID.String()})
}

func DebugState(c *fiber.Ctx) error {

	Matcher.mu.Lock()
	queues := make(map[string]map[string][]string)
	for cat, roles := range Matcher.queues {
		queues[cat] = make(map[string][]string)
		for role, list := range roles {
			for _, e := range list {
				queues[cat][role] = append(queues[cat][role], e.UserID.String())
			}
		}
	}
	ttl := Matcher.ttl.String()
	Matcher.mu.Unlock()

	active := utils.DefaultNotifier.ActiveUserIDs()
	activeStr := make([]string, 0, len(active))
	for _, id := range active {
		activeStr = append(activeStr, id.String())
	}

	resp := fiber.Map{
		"queues":       queues,
		"active_users": activeStr,
		"matcher_ttl":  ttl,
		"message_chan": len(messageChan),
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func DebugNotify(c *fiber.Ctx) error {
	var payload struct {
		UserID string                 `json:"user_id"`
		Body   map[string]interface{} `json:"body"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
	}
	uid, err := uuid.Parse(payload.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
	}
	if payload.Body == nil {
		payload.Body = map[string]interface{}{"event": "debug", "ts": time.Now().UTC().String()}
	}
	if err := utils.DefaultNotifier.Send(uid, payload.Body); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to send", "detail": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"sent": true})
}
