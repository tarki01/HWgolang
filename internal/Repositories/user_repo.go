package repository

import (
	"database/sql"
	"errors"

	models "HWGO/internal/Core"

	"github.com/lib/pq"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(username, email, passwordHash string) (*models.User, error) {
	query := `INSERT INTO users (username, email, password_hash)
	          VALUES ($1, $2, $3)
	          RETURNING id, username, email, created_at`

	user := &models.User{}
	err := r.db.QueryRow(query, username, email, passwordHash).
		Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, errors.New("user with this email or username already exists")
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, email).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *UserRepository) FindByID(id int) (*models.User, error) {
	query := `SELECT id, username, email, password_hash, created_at FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(query, id).
		Scan(&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("user not found")
	}
	return user, err
}

func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`, email).Scan(&exists)
	return exists, err
}

func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`, username).Scan(&exists)
	return exists, err
}
