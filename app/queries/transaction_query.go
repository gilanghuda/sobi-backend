package queries

import (
	"database/sql"
	"errors"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/google/uuid"
)

type TransactionQueries struct {
	DB *sql.DB
}

func (q *TransactionQueries) CreateTransaction(t *models.Transaction) error {
	query := `INSERT INTO transactions (id, user_id, ahli_id, amount, status, payment_url, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := q.DB.Exec(query, t.ID, t.UserID, t.AhliID, t.Amount, t.Status, t.PaymentURL, t.CreatedAt, t.UpdatedAt)
	if err != nil {
		return errors.New("unable to create transaction")
	}
	return nil
}

func (q *TransactionQueries) GetTransactionByID(id uuid.UUID) (models.Transaction, error) {
	t := models.Transaction{}
	query := `SELECT id, user_id, ahli_id, amount, status, payment_url, created_at, updated_at FROM transactions WHERE id = $1`
	err := q.DB.QueryRow(query, id).Scan(&t.ID, &t.UserID, &t.AhliID, &t.Amount, &t.Status, &t.PaymentURL, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return t, errors.New("transaction not found")
		}
		return t, errors.New("unable to get transaction")
	}
	return t, nil
}

func (q *TransactionQueries) UpdateTransactionStatus(id uuid.UUID, status string) error {
	query := `UPDATE transactions SET status = $2, updated_at = now() WHERE id = $1`
	res, err := q.DB.Exec(query, id, status)
	if err != nil {
		return errors.New("unable to update transaction")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return errors.New("unable to update transaction")
	}
	if rows == 0 {
		return errors.New("transaction not found")
	}
	return nil
}
