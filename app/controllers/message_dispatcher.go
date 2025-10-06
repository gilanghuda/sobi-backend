package controllers

import (
	"encoding/json"
	"log"

	"github.com/gilanghuda/sobi-backend/app/queries"
	"github.com/gilanghuda/sobi-backend/pkg/database"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
)

func StartMessageDispatcher() {
	go func() {
		for msg := range messageChan {
			payload := map[string]interface{}{
				"event":      "message",
				"id":         msg.ID,
				"room_id":    msg.RoomID,
				"user_id":    msg.UserID,
				"text":       msg.Text,
				"visible":    msg.Visible,
				"created_at": msg.CreatedAt,
			}
			data, err := json.Marshal(payload)
			if err != nil {
				log.Printf("dispatcher: marshal error: %v", err)
				continue
			}

			q := queries.ChatQueries{DB: database.DB}
			room, err := q.GetRoomByID(msg.RoomID)
			if err != nil {
				log.Printf("dispatcher: failed to get room %v: %v", msg.RoomID, err)
				// fallback: notify message sender
				sendPayloadToUser(msg.UserID, data)
				continue
			}

			sendPayloadToUser(room.OwnerID, data)
			if room.TargetID != nil && *room.TargetID != room.OwnerID {
				sendPayloadToUser(*room.TargetID, data)
			}
		}
	}()
}

func sendPayloadToUser(uid uuid.UUID, data []byte) {
	clientsMu.RLock()
	connsMap, ok := clientsByUser[uid]
	if !ok || len(connsMap) == 0 {
		clientsMu.RUnlock()
		return
	}
	conns := make([]*client, 0, len(connsMap))
	for c := range connsMap {
		conns = append(conns, c)
	}
	clientsMu.RUnlock()

	for _, c := range conns {
		if c == nil || c.conn == nil {
			continue
		}
		if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Printf("dispatcher: write error to user %v: %v", uid, err)
			clientsMu.Lock()
			if m, ok := clientsByUser[uid]; ok {
				delete(m, c)
				if len(m) == 0 {
					delete(clientsByUser, uid)
				}
			}
			clientsMu.Unlock()
			_ = c.conn.Close()
		}
	}
}
