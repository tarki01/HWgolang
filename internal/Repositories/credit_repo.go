package repository

import (
	"database/sql"
	"errors"
	"time"

	models "HWGO/internal/Core"
)

type CreditRepository struct {
	db *sql.DB
}

func NewCreditRepository(db *sql.DB) *CreditRepository {
	return &CreditRepository{db: db}
}

func (r *CreditRepository) Create(tx *sql.Tx, userID, accountID int, principal, rate float64, termMonths int, monthly float64) (*models.Credit, error) {
	query := `INSERT INTO credits (user_id, account_id, principal, interest_rate, term_months, monthly_payment, status)
	          VALUES ($1, $2, $3, $4, $5, $6, 'active')
	          RETURNING id, user_id, account_id, principal, interest_rate, term_months, monthly_payment, status, created_at`

	c := &models.Credit{}
	var err error
	if tx != nil {
		err = tx.QueryRow(query, userID, accountID, principal, rate, termMonths, monthly).
			Scan(&c.ID, &c.UserID, &c.AccountID, &c.Principal, &c.InterestRate, &c.TermMonths, &c.MonthlyPayment, &c.Status, &c.CreatedAt)
	} else {
		err = r.db.QueryRow(query, userID, accountID, principal, rate, termMonths, monthly).
			Scan(&c.ID, &c.UserID, &c.AccountID, &c.Principal, &c.InterestRate, &c.TermMonths, &c.MonthlyPayment, &c.Status, &c.CreatedAt)
	}
	return c, err
}

func (r *CreditRepository) FindByID(id int) (*models.Credit, error) {
	query := `SELECT id, user_id, account_id, principal, interest_rate, term_months, monthly_payment, status, created_at
	          FROM credits WHERE id = $1`

	c := &models.Credit{}
	err := r.db.QueryRow(query, id).
		Scan(&c.ID, &c.UserID, &c.AccountID, &c.Principal, &c.InterestRate, &c.TermMonths, &c.MonthlyPayment, &c.Status, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, errors.New("credit not found")
	}
	return c, err
}

func (r *CreditRepository) FindByUserID(userID int) ([]models.Credit, error) {
	query := `SELECT id, user_id, account_id, principal, interest_rate, term_months, monthly_payment, status, created_at
	          FROM credits WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var credits []models.Credit
	for rows.Next() {
		var c models.Credit
		if err := rows.Scan(&c.ID, &c.UserID, &c.AccountID, &c.Principal, &c.InterestRate, &c.TermMonths, &c.MonthlyPayment, &c.Status, &c.CreatedAt); err != nil {
			return nil, err
		}
		credits = append(credits, c)
	}
	return credits, rows.Err()
}

func (r *CreditRepository) UpdateStatus(id int, status string) error {
	_, err := r.db.Exec(`UPDATE credits SET status = $1 WHERE id = $2`, status, id)
	return err
}

func (r *CreditRepository) GetOverdueSchedules() ([]models.PaymentSchedule, error) {
	query := `SELECT ps.id, ps.credit_id, ps.payment_date, ps.amount, ps.principal, ps.interest, ps.status, ps.paid_at
	          FROM payment_schedules ps
	          WHERE ps.status = 'pending' AND ps.payment_date < $1`

	rows, err := r.db.Query(query, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedules []models.PaymentSchedule
	for rows.Next() {
		var s models.PaymentSchedule
		if err := rows.Scan(&s.ID, &s.CreditID, &s.PaymentDate, &s.Amount, &s.Principal, &s.Interest, &s.Status, &s.PaidAt); err != nil {
			return nil, err
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

func (r *CreditRepository) GetDB() *sql.DB {
	return r.db
}
