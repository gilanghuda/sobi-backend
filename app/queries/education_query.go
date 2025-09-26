package queries

import (
	"database/sql"
	"time"

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
