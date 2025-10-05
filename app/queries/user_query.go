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

	query := `SELECT uid, username, user_role, email, phone_number, gender, avatar, password_hash, verified,  created_at, updated_at
			  FROM users WHERE uid = $1`

	err := q.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.UserRole,
		&user.Email,
		&user.PhoneNumber,
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

// UpdateOTPByEmail updates the OTP for a user identified by email
func (q *UserQueries) UpdateOTPByEmail(email string, otp string) error {
	query := `UPDATE users SET otp = $1, updated_at = now() WHERE email = $2`
	res, err := q.DB.Exec(query, otp, email)
	if err != nil {
		return errors.New("unable to update otp, DB error")
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("no user updated")
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

func (q *UserQueries) GetUsersByRole(role string) ([]models.User, error) {
	users := []models.User{}
	query := `SELECT uid, username, user_role, email, phone_number, gender, avatar, password_hash, verified, created_at, updated_at FROM users WHERE user_role = $1`
	rows, err := q.DB.Query(query, role)
	if err != nil {
		return users, errors.New("unable to get users, DB error")
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.UserRole,
			&user.Email,
			&user.PhoneNumber,
			&user.Gender,
			&user.Avatar,
			&user.PasswordHash,
			&user.Verified,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return users, errors.New("error scanning user row")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return users, errors.New("error iterating user rows")
	}

	return users, nil
}

// CreateAhli promotes an existing user to role 'ahli' and inserts ahli-specific data into ahli
func (q *UserQueries) CreateAhli(uid uuid.UUID, req *models.PromoteAhliRequest) error {
	tx, err := q.DB.Begin()
	if err != nil {
		return errors.New("unable to start transaction")
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// update user role to 'ahli'
	_, err = tx.Exec(`UPDATE users SET user_role = 'ahli', updated_at = now() WHERE uid = $1`, uid)
	if err != nil {
		tx.Rollback()
		return errors.New("unable to update user role, DB error")
	}

	// prepare open_time parameter (use nil when empty)
	var openTimeParam interface{}
	if req.OpenTime == "" {
		openTimeParam = nil
	} else {
		openTimeParam = req.OpenTime
	}

	// insert ahli record into ahli
	_, err = tx.Exec(`INSERT INTO ahli (uid, price, category, open_time, rating) VALUES ($1, $2, $3, $4, $5)`,
		uid, req.Price, req.Category, openTimeParam, req.Rating,
	)
	if err != nil {
		println(err.Error())
		tx.Rollback()
		return errors.New("unable to create ahli, DB error")
	}

	if err := tx.Commit(); err != nil {
		return errors.New("unable to commit transaction")
	}

	return nil
}

// GetAhliUsers returns users that are ahli along with ahli-specific fields
func (q *UserQueries) GetAhliUsers() ([]models.User, error) {
	users := []models.User{}
	query := `SELECT u.uid, u.username, u.user_role, u.email, u.phone_number, u.gender, u.avatar, u.password_hash, u.verified, u.created_at, u.updated_at,
	a.price, a.category, a.open_time, a.rating
	FROM users u JOIN ahli a ON u.uid = a.uid WHERE u.user_role = 'ahli'`

	rows, err := q.DB.Query(query)
	if err != nil {
		return users, errors.New("unable to get ahli users, DB error")
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		var openTime sql.NullTime
		if err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.UserRole,
			&user.Email,
			&user.PhoneNumber,
			&user.Gender,
			&user.Avatar,
			&user.PasswordHash,
			&user.Verified,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Price,
			&user.Category,
			&openTime,
			&user.Rating,
		); err != nil {
			return users, errors.New("error scanning ahli user row")
		}
		if openTime.Valid {
			user.OpenTime = openTime.Time.Format("15:04:05")
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return users, errors.New("error iterating ahli user rows")
	}

	return users, nil
}

func (q *UserQueries) GetUserByUsername(username string) (models.User, error) {
	user := models.User{}

	query := `SELECT uid, username, user_role, email, password_hash, verified, created_at, updated_at
			  FROM users WHERE username = $1`

	err := q.DB.QueryRow(query, username).Scan(
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
