package utils

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

// Notifier manages active WebSocket connections and sending notifications.
type Notifier struct {
	mu    sync.RWMutex
	conns map[uuid.UUID]*websocket.Conn
}

// DefaultNotifier is the package-level notifier instance.
var DefaultNotifier = NewNotifier()

// NewNotifier creates a new Notifier.
func NewNotifier() *Notifier {
	return &Notifier{
		conns: make(map[uuid.UUID]*websocket.Conn),
	}
}

// Register registers a websocket connection for a user.
func (n *Notifier) Register(userID uuid.UUID, conn *websocket.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.conns[userID] = conn
	log.Printf("event=ws_register user=%s total_connections=%d", userID.String(), len(n.conns))
}

// Unregister removes the websocket connection for a user.
func (n *Notifier) Unregister(userID uuid.UUID) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if conn, ok := n.conns[userID]; ok {
		_ = conn.Close()
		delete(n.conns, userID)
	}
	log.Printf("event=ws_unregister user=%s total_connections=%d", userID.String(), len(n.conns))
}

// Send sends a JSON-serializable payload to the user's websocket connection.
func (n *Notifier) Send(userID uuid.UUID, payload interface{}) error {
	n.mu.RLock()
	conn, ok := n.conns[userID]
	n.mu.RUnlock()
	if !ok || conn == nil {
		log.Printf("event=notify_skip user=%s reason=no_connection", userID.String())
		return ErrNoConnection
	}

	msg, err := json.Marshal(payload)
	if err != nil {
		log.Printf("event=notify_error user=%s error=%v", userID.String(), err)
		return err
	}

	// log payload string for debug
	log.Printf("event=notify_send user=%s payload=%s", userID.String(), string(msg))

	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Printf("event=notify_error_write user=%s error=%v", userID.String(), err)
		return err
	}

	log.Printf("event=notify_sent user=%s payload_len=%d", userID.String(), len(msg))
	return nil
}

// ActiveUserIDs returns a snapshot of currently connected user IDs.
func (n *Notifier) ActiveUserIDs() []uuid.UUID {
	n.mu.RLock()
	defer n.mu.RUnlock()
	out := make([]uuid.UUID, 0, len(n.conns))
	for id := range n.conns {
		out = append(out, id)
	}
	return out
}

// ErrNoConnection is returned when there is no websocket connection for the user.
var ErrNoConnection = &NoConnError{}

type NoConnError struct{}

func (e *NoConnError) Error() string { return "no websocket connection for user" }
