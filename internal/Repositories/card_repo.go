package repository

import (
	"database/sql"
	"errors"

	models "HWGO/internal/Core"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) Create(accountID int, encNum, encExpiry, cvvHash, hmacNum, lastFour string) (*models.Card, error) {
	query := `INSERT INTO cards (account_id, encrypted_number, encrypted_expiry, cvv_hash, hmac_number, last_four)
	          VALUES ($1, $2, $3, $4, $5, $6)
	          RETURNING id, account_id, encrypted_number, encrypted_expiry, cvv_hash, hmac_number, last_four, created_at`

	c := &models.Card{}
	err := r.db.QueryRow(query, accountID, encNum, encExpiry, cvvHash, hmacNum, lastFour).
		Scan(&c.ID, &c.AccountID, &c.EncryptedNumber, &c.EncryptedExpiry, &c.CVVHash, &c.HMACNumber, &c.LastFour, &c.CreatedAt)
	return c, err
}

func (r *CardRepository) FindByAccountID(accountID int) ([]models.Card, error) {
	query := `SELECT id, account_id, encrypted_number, encrypted_expiry, cvv_hash, hmac_number, last_four, created_at
	          FROM cards WHERE account_id = $1`

	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.Card
	for rows.Next() {
		var c models.Card
		if err := rows.Scan(&c.ID, &c.AccountID, &c.EncryptedNumber, &c.EncryptedExpiry, &c.CVVHash, &c.HMACNumber, &c.LastFour, &c.CreatedAt); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

func (r *CardRepository) FindByID(id int) (*models.Card, error) {
	query := `SELECT id, account_id, encrypted_number, encrypted_expiry, cvv_hash, hmac_number, last_four, created_at
	          FROM cards WHERE id = $1`

	c := &models.Card{}
	err := r.db.QueryRow(query, id).
		Scan(&c.ID, &c.AccountID, &c.EncryptedNumber, &c.EncryptedExpiry, &c.CVVHash, &c.HMACNumber, &c.LastFour, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("card not found")
	}
	return c, err
}
