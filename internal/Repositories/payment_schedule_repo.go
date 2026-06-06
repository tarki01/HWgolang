package repository

import (
	models "HWGO/internal/Core"
	"database/sql"
	"time"
)

type PaymentScheduleRepository struct {
	db *sql.DB
}

func NewPaymentScheduleRepository(db *sql.DB) *PaymentScheduleRepository {
	return &PaymentScheduleRepository{db: db}
}

func (r *PaymentScheduleRepository) CreateBatch(tx *sql.Tx, schedules []models.PaymentSchedule) error {
	query := `INSERT INTO payment_schedules (credit_id, payment_date, amount, principal, interest, status)
	          VALUES ($1, $2, $3, $4, $5, 'pending')`

	for _, s := range schedules {
		var err error
		if tx != nil {
			_, err = tx.Exec(query, s.CreditID, s.PaymentDate, s.Amount, s.Principal, s.Interest)
		} else {
			_, err = r.db.Exec(query, s.CreditID, s.PaymentDate, s.Amount, s.Principal, s.Interest)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PaymentScheduleRepository) FindByCreditID(creditID int) ([]models.PaymentSchedule, error) {
	query := `SELECT id, credit_id, payment_date, amount, principal, interest, status, paid_at
	          FROM payment_schedules WHERE credit_id = $1 ORDER BY payment_date`

	rows, err := r.db.Query(query, creditID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PaymentSchedule
	for rows.Next() {
		var s models.PaymentSchedule
		if err := rows.Scan(&s.ID, &s.CreditID, &s.PaymentDate, &s.Amount, &s.Principal, &s.Interest, &s.Status, &s.PaidAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}

func (r *PaymentScheduleRepository) MarkPaid(tx *sql.Tx, id int) error {
	now := time.Now()
	query := `UPDATE payment_schedules SET status = 'paid', paid_at = $1 WHERE id = $2`
	var err error
	if tx != nil {
		_, err = tx.Exec(query, now, id)
	} else {
		_, err = r.db.Exec(query, now, id)
	}
	return err
}

func (r *PaymentScheduleRepository) MarkOverdue(id int) error {
	_, err := r.db.Exec(`UPDATE payment_schedules SET status = 'overdue' WHERE id = $1`, id)
	return err
}

func (r *PaymentScheduleRepository) UpdateAmount(id int, newAmount float64) error {
	_, err := r.db.Exec(`UPDATE payment_schedules SET amount = $1 WHERE id = $2`, newAmount, id)
	return err
}

func (r *PaymentScheduleRepository) GetUpcoming(accountID int, until time.Time) ([]models.PaymentSchedule, error) {
	query := `SELECT ps.id, ps.credit_id, ps.payment_date, ps.amount, ps.principal, ps.interest, ps.status, ps.paid_at
	          FROM payment_schedules ps
	          JOIN credits c ON c.id = ps.credit_id
	          WHERE c.account_id = $1 AND ps.status = 'pending' AND ps.payment_date <= $2
	          ORDER BY ps.payment_date`

	rows, err := r.db.Query(query, accountID, until)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.PaymentSchedule
	for rows.Next() {
		var s models.PaymentSchedule
		if err := rows.Scan(&s.ID, &s.CreditID, &s.PaymentDate, &s.Amount, &s.Principal, &s.Interest, &s.Status, &s.PaidAt); err != nil {
			return nil, err
		}
		items = append(items, s)
	}
	return items, rows.Err()
}
