package queries

import (
	"database/sql"
	"errors"
	"sort"
	"time"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/google/uuid"
)

type ChatQueries struct {
	DB *sql.DB
}

func (q *ChatQueries) CreateRoom(r *models.Room) error {
	query := `INSERT INTO rooms (id, owner_id, target_id, category, visible, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)`
	_, err := q.DB.Exec(query, r.ID, r.OwnerID, r.TargetID, r.Category, r.Visible, r.CreatedAt, r.UpdatedAt)
	if err != nil {
		return errors.New("unable to create room")
	}
	return nil
}

func (q *ChatQueries) GetRoomsByUser(userID uuid.UUID) ([]models.Room, error) {
	var res []models.Room
	query := `SELECT id, owner_id, target_id, category, visible, created_at, updated_at FROM rooms WHERE owner_id = $1 OR target_id = $1`
	rows, err := q.DB.Query(query, userID)
	if err != nil {
		return res, errors.New("unable to query rooms")
	}
	defer rows.Close()
	for rows.Next() {
		var r models.Room
		var target sql.NullString
		if err := rows.Scan(&r.ID, &r.OwnerID, &target, &r.Category, &r.Visible, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return res, err
		}
		if target.Valid {
			uid, err := uuid.Parse(target.String)
			if err == nil {
				r.TargetID = &uid
			}
		}
		res = append(res, r)
	}
	return res, nil
}

func (q *ChatQueries) GetRoomByID(id uuid.UUID) (models.Room, error) {
	r := models.Room{}
	query := `SELECT id, owner_id, target_id, category, visible, created_at, updated_at FROM rooms WHERE id = $1`
	var target sql.NullString
	if err := q.DB.QueryRow(query, id).Scan(&r.ID, &r.OwnerID, &target, &r.Category, &r.Visible, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return r, errors.New("room not found")
		}
		return r, errors.New("unable to get room")
	}
	if target.Valid {
		uid, err := uuid.Parse(target.String)
		if err == nil {
			r.TargetID = &uid
		}
	}
	return r, nil
}

func (q *ChatQueries) CreateMessage(m *models.Message) error {
	query := `INSERT INTO messages (id, room_id, user_id, text, visible, created_at) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := q.DB.Exec(query, m.ID, m.RoomID, m.UserID, m.Text, m.Visible, m.CreatedAt)
	if err != nil {
		return errors.New("unable to create message")
	}
	return nil
}

func (q *ChatQueries) GetMessagesByRoom(roomID uuid.UUID, limit int) ([]models.Message, error) {
	var res []models.Message
	query := `SELECT id, room_id, user_id, text, visible, created_at FROM messages WHERE room_id = $1 ORDER BY created_at ASC LIMIT $2`
	rows, err := q.DB.Query(query, roomID, limit)
	if err != nil {
		return res, errors.New("unable to query messages")
	}
	defer rows.Close()
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.RoomID, &m.UserID, &m.Text, &m.Visible, &m.CreatedAt); err != nil {
			return res, err
		}
		res = append(res, m)
	}
	return res, nil
}

func (q *ChatQueries) GetRecentChats(userID uuid.UUID, limit int) ([]models.RecentChat, error) {
	rows, err := q.DB.Query(`
	SELECT r.id, r.owner_id, r.target_id, m.text, m.created_at
	FROM rooms r
	JOIN LATERAL (
	  SELECT text, created_at FROM messages WHERE room_id = r.id ORDER BY created_at DESC LIMIT 1
	) m ON true
	WHERE r.owner_id = $1 OR r.target_id = $1
	ORDER BY m.created_at DESC
	LIMIT $2
	`, userID, limit*5)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recentMap := make(map[uuid.UUID]models.RecentChat)
	for rows.Next() {
		var roomID uuid.UUID
		var owner uuid.UUID
		var target sql.NullString
		var text sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&roomID, &owner, &target, &text, &createdAt); err != nil {
			return nil, err
		}
		var other uuid.UUID
		if target.Valid {
			uid, _ := uuid.Parse(target.String)
			if owner == userID {
				other = uid
			} else {
				other = owner
			}
		} else {
			if owner == userID {
				other = uuid.Nil
			} else {
				other = owner
			}
		}

		if other == uuid.Nil || other == userID {
			continue
		}

		rc, ok := recentMap[other]
		if !ok || createdAt.After(rc.LastAt) {
			recentMap[other] = models.RecentChat{OtherUserID: other, RoomID: roomID, LastMessage: text.String, LastAt: createdAt}
		}
	}

	out := make([]models.RecentChat, 0, len(recentMap))
	for _, v := range recentMap {
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].LastAt.After(out[j].LastAt) })

	if len(out) > limit {
		out = out[:limit]
	}

	return out, nil
}

func (q *ChatQueries) GetRecentChatsAsTarget(userID uuid.UUID, limit int) ([]models.RecentChat, error) {
	rows, err := q.DB.Query(`
	SELECT r.id, r.owner_id, m.text, m.created_at
	FROM rooms r
	JOIN LATERAL (
	  SELECT text, created_at FROM messages WHERE room_id = r.id ORDER BY created_at DESC LIMIT 1
	) m ON true
	WHERE r.target_id = $1
	ORDER BY m.created_at DESC
	LIMIT $2
	`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.RecentChat
	for rows.Next() {
		var roomID uuid.UUID
		var owner uuid.UUID
		var text sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&roomID, &owner, &text, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, models.RecentChat{OtherUserID: owner, RoomID: roomID, LastMessage: text.String, LastAt: createdAt})
	}
	return out, nil
}

func (q *ChatQueries) GetActiveRoom(userID uuid.UUID, startTime, endTime time.Time) (models.Room, error) {
	r := models.Room{}
	query := `SELECT id, owner_id, target_id, category, visible, created_at, updated_at FROM rooms WHERE (owner_id = $1 OR target_id = $1) AND created_at >= $2 AND created_at <= $3 ORDER BY created_at DESC LIMIT 1`
	var target sql.NullString
	if err := q.DB.QueryRow(query, userID, startTime, endTime).Scan(&r.ID, &r.OwnerID, &target, &r.Category, &r.Visible, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return r, errors.New("no active room")
		}
		return r, errors.New("unable to get active room")
	}
	if target.Valid {
		uid, err := uuid.Parse(target.String)
		if err == nil {
			r.TargetID = &uid
		}
	}
	return r, nil
}
