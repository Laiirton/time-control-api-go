package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/Laiirton/time-control-api-go/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (name, type, email, password, role, department, phone, location, shift, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	return r.db.QueryRow(
		query,
		user.Name, user.Type, user.Email, user.Password,
		user.Role, user.Department, user.Phone, user.Location, user.Shift,
		user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, type, email, password, role, department, phone, location, shift, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Type, &user.Email, &user.Password,
		&user.Role, &user.Department, &user.Phone, &user.Location, &user.Shift,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id int64) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, name, type, email, password, role, department, phone, location, shift, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Type, &user.Email, &user.Password,
		&user.Role, &user.Department, &user.Phone, &user.Location, &user.Shift,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindAll(page, limit int) ([]models.User, int, error) {
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	query := `
		SELECT id, name, type, email, password, role, department, phone, location, shift, created_at, updated_at
		FROM users ORDER BY id LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(
			&u.ID, &u.Name, &u.Type, &u.Email, &u.Password,
			&u.Role, &u.Department, &u.Phone, &u.Location, &u.Shift,
			&u.CreatedAt, &u.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, nil
}

func (r *UserRepository) Update(id int64, req *models.UpdateUserRequest) (*models.User, error) {
	sets := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, *req.Name)
		argIdx++
	}
	if req.Type != nil {
		sets = append(sets, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *req.Type)
		argIdx++
	}
	if req.Email != nil {
		sets = append(sets, fmt.Sprintf("email = $%d", argIdx))
		args = append(args, *req.Email)
		argIdx++
	}
	if req.Password != nil {
		sets = append(sets, fmt.Sprintf("password = $%d", argIdx))
		args = append(args, *req.Password)
		argIdx++
	}
	if req.Role != nil {
		sets = append(sets, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *req.Role)
		argIdx++
	}
	if req.Department != nil {
		sets = append(sets, fmt.Sprintf("department = $%d", argIdx))
		args = append(args, *req.Department)
		argIdx++
	}
	if req.Phone != nil {
		sets = append(sets, fmt.Sprintf("phone = $%d", argIdx))
		args = append(args, *req.Phone)
		argIdx++
	}
	if req.Location != nil {
		sets = append(sets, fmt.Sprintf("location = $%d", argIdx))
		args = append(args, *req.Location)
		argIdx++
	}
	if req.Shift != nil {
		sets = append(sets, fmt.Sprintf("shift = $%d", argIdx))
		args = append(args, *req.Shift)
		argIdx++
	}

	if len(sets) == 0 {
		return r.FindByID(id)
	}

	sets = append(sets, fmt.Sprintf("updated_at = $%d", argIdx))
	args = append(args, time.Now())
	argIdx++

	args = append(args, id)
	query := fmt.Sprintf(
		"UPDATE users SET %s WHERE id = $%d RETURNING id, name, type, email, password, role, department, phone, location, shift, created_at, updated_at",
		strings.Join(sets, ", "), argIdx,
	)

	user := &models.User{}
	err := r.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Name, &user.Type, &user.Email, &user.Password,
		&user.Role, &user.Department, &user.Phone, &user.Location, &user.Shift,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
