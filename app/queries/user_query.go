package queries

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gilanghuda/sobi-backend/app/models"
	"github.com/google/uuid"
)

type UserQueries struct {
	DB *sql.DB
}

func (q *UserQueries) GetUserByID(id uuid.UUID) (models.User, error) {
	user := models.User{}

	query := `SELECT uid, username, user_role, email, gender, avatar, password_hash, verified,  created_at, updated_at
			  FROM users WHERE uid = $1`

	err := q.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.UserRole,
		&user.Email,
		&user.Gender,
		&user.Avatar,
		&user.PasswordHash,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		println(err.Error())
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		return user, errors.New("unable to get user, DB error")
	}

	return user, nil
}

func (q *UserQueries) GetUserByEmail(email string) (models.User, error) {
	user := models.User{}

	query := `SELECT uid, username, user_role, email, password_hash, verified, created_at, updated_at
			  FROM users WHERE email = $1`

	err := q.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.UserRole,
		&user.Email,
		&user.PasswordHash,
		&user.Verified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, errors.New("user not found")
		}
		println(err.Error())
		return user, errors.New("unable to get user, DB error")
	}

	return user, nil
}

func (q *UserQueries) CreateUser(u *models.User) error {
	query := `INSERT INTO users (uid, username, user_role, email, password_hash, phone_number, verified, created_at, updated_at, otp, gender, avatar)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`

	_, err := q.DB.Exec(query,
		u.ID,
		u.Username,
		u.UserRole,
		u.Email,
		u.PasswordHash,
		u.PhoneNumber,
		u.Verified,
		u.CreatedAt,
		u.UpdatedAt,
		u.OTP,
		u.Gender,
		u.Avatar,
	)

	if err != nil {
		println("disini kan?" + err.Error())
		return errors.New("unable to create user, DB error")
	}

	return nil
}

func (q *UserQueries) VerifyOTPByEmail(email string, otp string) error {
	query := `UPDATE users SET verified = TRUE,  updated_at = now() WHERE email = $1 AND otp = $2 AND verified = FALSE`
	res, err := q.DB.Exec(query, email, otp)
	if err != nil {
		return errors.New("unable to verify OTP, DB error")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("invalid otp or already verified")
	}
	return nil
}

func (q *UserQueries) UpdateUser(userID uuid.UUID, req *models.UpdateUserRequest) error {
	setClauses := []string{}
	args := []interface{}{}
	argID := 1

	if req.Username != nil {
		setClauses = append(setClauses, fmt.Sprintf("username = $%d", argID))
		args = append(args, *req.Username)
		argID++
	}
	if req.PhoneNumber != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone_number = $%d", argID))
		args = append(args, *req.PhoneNumber)
		argID++
	}
	if req.Gender != nil {
		setClauses = append(setClauses, fmt.Sprintf("gender = $%d", argID))
		args = append(args, *req.Gender)
		argID++
	}
	if req.Avatar != nil {
		setClauses = append(setClauses, fmt.Sprintf("avatar = $%d", argID))
		args = append(args, *req.Avatar)
		argID++
	}

	if len(setClauses) == 0 {
		return errors.New("no fields to update")
	}

	setClauses = append(setClauses, "updated_at = now()")
	query := fmt.Sprintf(`UPDATE users SET %s WHERE uid = $%d`, strings.Join(setClauses, ", "), argID)

	args = append(args, userID)

	res, err := q.DB.Exec(query, args...)
	if err != nil {
		return errors.New("unable to update user, DB error")
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("no user updated")
	}
	return nil
}

func (q *UserQueries) DeleteUser(id uuid.UUID) error {
	query := `DELETE FROM users WHERE uid = $1`

	res, err := q.DB.Exec(query, id)
	if err != nil {
		return errors.New("unable to delete user, DB error")
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no user deleted")
	}

	return nil
}
