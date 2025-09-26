package queries

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/gilanghuda/sobi-backend/app/models"
)

// GetAllEducations retrieves all educations ordered by created_at desc
func GetAllEducations(db *sql.DB) ([]models.Education, error) {
	query := `SELECT id, title, subtitle, video_url, duration, author, description, created_at FROM educations ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eds []models.Education
	for rows.Next() {
		var e models.Education
		var createdAt time.Time
		if err := rows.Scan(&e.ID, &e.Title, &e.Subtitle, &e.VideoURL, &e.Duration, &e.Author, &e.Description, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt = createdAt
		eds = append(eds, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return eds, nil
}

// GetEducationByID retrieves one education by id
func GetEducationByID(db *sql.DB, id string) (*models.Education, error) {
	query := `SELECT id, title, subtitle, video_url, duration, author, description, created_at FROM educations WHERE id = $1 LIMIT 1`
	row := db.QueryRow(query, id)
	var e models.Education
	var createdAt time.Time
	if err := row.Scan(&e.ID, &e.Title, &e.Subtitle, &e.VideoURL, &e.Duration, &e.Author, &e.Description, &createdAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	e.CreatedAt = createdAt
	return &e, nil
}

// InsertHistoryEducation inserts a history record for a user and education; ignores duplicates
func InsertHistoryEducation(db *sql.DB, h models.HistoryEducation) error {
	query := `INSERT INTO history_education (user_id, education_id) VALUES ($1, $2) ON CONFLICT (user_id, education_id) DO NOTHING`
	_, err := db.Exec(query, h.UserID, h.EducationID)
	return err
}

// GetHistoryByUser returns educations that the user has viewed, ordered by education created_at desc
func GetHistoryByUser(db *sql.DB, userID uuid.UUID) ([]models.Education, error) {
	query := `SELECT e.id, e.title, e.subtitle, e.video_url, e.duration, e.author, e.description, e.created_at FROM history_education h JOIN educations e ON e.id = h.education_id WHERE h.user_id = $1 ORDER BY e.created_at DESC`
	rows, err := db.Query(query, userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eds []models.Education
	for rows.Next() {
		var e models.Education
		var createdAt time.Time
		if err := rows.Scan(&e.ID, &e.Title, &e.Subtitle, &e.VideoURL, &e.Duration, &e.Author, &e.Description, &createdAt); err != nil {
			return nil, err
		}
		e.CreatedAt = createdAt
		eds = append(eds, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return eds, nil
}
