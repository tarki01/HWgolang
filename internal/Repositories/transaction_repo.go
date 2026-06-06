package repository

import (
	"database/sql"
	"time"

	models "HWGO/internal/Core"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(tx *sql.Tx, fromID, toID *int, amount float64, txType, desc string) (*models.Transaction, error) {
	query := `INSERT INTO transactions (from_account_id, to_account_id, amount, transaction_type, description)
	          VALUES ($1, $2, $3, $4, $5)
	          RETURNING id, from_account_id, to_account_id, amount, transaction_type, description, created_at`

	t := &models.Transaction{}
	var err error
	if tx != nil {
		err = tx.QueryRow(query, fromID, toID, amount, txType, desc).
			Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.TransactionType, &t.Description, &t.CreatedAt)
	} else {
		err = r.db.QueryRow(query, fromID, toID, amount, txType, desc).
			Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.TransactionType, &t.Description, &t.CreatedAt)
	}
	return t, err
}

func (r *TransactionRepository) FindByAccountID(accountID int) ([]models.Transaction, error) {
	query := `SELECT id, from_account_id, to_account_id, amount, transaction_type, description, created_at
	          FROM transactions
	          WHERE from_account_id = $1 OR to_account_id = $1
	          ORDER BY created_at DESC`

	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.FromAccountID, &t.ToAccountID, &t.Amount, &t.TransactionType, &t.Description, &t.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	return txs, rows.Err()
}

func (r *TransactionRepository) GetMonthlyStats(accountID int, from, to time.Time) (income, expenses float64, err error) {
	query := `SELECT
	    COALESCE(SUM(CASE WHEN to_account_id = $1 THEN amount ELSE 0 END), 0),
	    COALESCE(SUM(CASE WHEN from_account_id = $1 THEN amount ELSE 0 END), 0)
	FROM transactions
	WHERE (from_account_id = $1 OR to_account_id = $1)
	  AND created_at >= $2 AND created_at < $3`

	err = r.db.QueryRow(query, accountID, from, to).Scan(&income, &expenses)
	return
}
