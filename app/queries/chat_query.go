package queries

import (
	"database/sql"
	"errors"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/google/uuid"
)

type ChatQueries struct {
	DB *sql.DB
}

func (q *ChatQueries) CreateRoom(r *models.Room) error {
	query := `INSERT INTO rooms (id, owner_id, category, visible, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6)`
	_, err := q.DB.Exec(query, r.ID, r.OwnerID, r.Category, r.Visible, r.CreatedAt, r.UpdatedAt)
	if err != nil {
		return errors.New("unable to create room")
	}
	return nil
}

func (q *ChatQueries) GetRoomsByUser(userID uuid.UUID) ([]models.Room, error) {
	var res []models.Room
	query := `SELECT id, owner_id, category, visible, created_at, updated_at FROM rooms WHERE owner_id = $1`
	rows, err := q.DB.Query(query, userID)
	if err != nil {
		return res, errors.New("unable to query rooms")
	}
	defer rows.Close()
	for rows.Next() {
		var r models.Room
		if err := rows.Scan(&r.ID, &r.OwnerID, &r.Category, &r.Visible, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return res, err
		}
		res = append(res, r)
	}
	return res, nil
}

func (q *ChatQueries) GetRoomByID(id uuid.UUID) (models.Room, error) {
	r := models.Room{}
	query := `SELECT id, owner_id, category, visible, created_at, updated_at FROM rooms WHERE id = $1`
	if err := q.DB.QueryRow(query, id).Scan(&r.ID, &r.OwnerID, &r.Category, &r.Visible, &r.CreatedAt, &r.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return r, errors.New("room not found")
		}
		return r, errors.New("unable to get room")
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
