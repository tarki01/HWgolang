package repository

import (
	"database/sql"
	"errors"

	models "HWGO/internal/Core"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(userID int, currency string) (*models.Account, error) {
	query := `INSERT INTO accounts (user_id, currency) VALUES ($1, $2)
	          RETURNING id, user_id, balance, currency, created_at`

	acc := &models.Account{}
	err := r.db.QueryRow(query, userID, currency).
		Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt)
	return acc, err
}

func (r *AccountRepository) FindByID(id int) (*models.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id = $1`

	acc := &models.Account{}
	err := r.db.QueryRow(query, id).
		Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("account not found")
	}
	return acc, err
}

func (r *AccountRepository) FindByUserID(userID int) ([]models.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE user_id = $1 ORDER BY created_at`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var acc models.Account
		if err := rows.Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, acc)
	}
	return accounts, rows.Err()
}

func (r *AccountRepository) UpdateBalance(tx *sql.Tx, accountID int, newBalance float64) error {
	query := `UPDATE accounts SET balance = $1 WHERE id = $2`
	var err error
	if tx != nil {
		_, err = tx.Exec(query, newBalance, accountID)
	} else {
		_, err = r.db.Exec(query, newBalance, accountID)
	}
	return err
}

func (r *AccountRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}

func (r *AccountRepository) FindByIDTx(tx *sql.Tx, id int) (*models.Account, error) {
	query := `SELECT id, user_id, balance, currency, created_at FROM accounts WHERE id = $1 FOR UPDATE`

	acc := &models.Account{}
	err := tx.QueryRow(query, id).
		Scan(&acc.ID, &acc.UserID, &acc.Balance, &acc.Currency, &acc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("account not found")
	}
	return acc, err
}
