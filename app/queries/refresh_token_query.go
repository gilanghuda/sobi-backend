package queries

import (
	"database/sql"
	"errors"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/google/uuid"
)

type RefreshTokenQueries struct {
	DB *sql.DB
}

func (q *RefreshTokenQueries) CreateRefreshToken(rt *models.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, revoked) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := q.DB.Exec(query, rt.ID, rt.UserID, rt.Token, rt.ExpiresAt, rt.CreatedAt, rt.Revoked)
	if err != nil {
		return errors.New("unable to create refresh token, DB error")
	}
	return nil
}

func (q *RefreshTokenQueries) GetRefreshTokenByToken(token string) (models.RefreshToken, error) {
	rt := models.RefreshToken{}
	query := `SELECT id, user_id, token, expires_at, created_at, revoked FROM refresh_tokens WHERE token = $1`
	err := q.DB.QueryRow(query, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return rt, errors.New("refresh token not found")
		}
		return rt, errors.New("unable to get refresh token, DB error")
	}
	return rt, nil
}

func (q *RefreshTokenQueries) RevokeRefreshToken(id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE id = $1`
	res, err := q.DB.Exec(query, id)
	if err != nil {
		return errors.New("unable to revoke refresh token, DB error")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no refresh token revoked")
	}
	return nil
}

func (q *RefreshTokenQueries) RevokeRefreshTokenByToken(token string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE token = $1`
	res, err := q.DB.Exec(query, token)
	if err != nil {
		return errors.New("unable to revoke refresh token by token, DB error")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no refresh token revoked")
	}
	return nil
}

func (q *RefreshTokenQueries) RevokeRefreshTokensByUser(userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE WHERE user_id = $1`
	_, err := q.DB.Exec(query, userID)
	if err != nil {
		return errors.New("unable to revoke refresh tokens for user, DB error")
	}
	return nil
}
