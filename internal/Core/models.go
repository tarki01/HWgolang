package models

import (
	"errors"
	"regexp"
	"time"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterRequest) Validate() error {
	if len(r.Username) < 3 || len(r.Username) > 50 {
		return errors.New("username must be between 3 and 50 characters")
	}
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}
	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *LoginRequest) Validate() error {
	if !emailRegex.MatchString(r.Email) {
		return errors.New("invalid email format")
	}
	if r.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type Account struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"created_at"`
}

type Card struct {
	ID              int       `json:"id"`
	AccountID       int       `json:"account_id"`
	EncryptedNumber string    `json:"-"`
	EncryptedExpiry string    `json:"-"`
	CVVHash         string    `json:"-"`
	HMACNumber      string    `json:"-"`
	LastFour        string    `json:"last_four"`
	CreatedAt       time.Time `json:"created_at"`
}

type CardView struct {
	ID        int    `json:"id"`
	AccountID int    `json:"account_id"`
	Number    string `json:"number"`
	Expiry    string `json:"expiry"`
	LastFour  string `json:"last_four"`
}

type Transaction struct {
	ID              int       `json:"id"`
	FromAccountID   *int      `json:"from_account_id,omitempty"`
	ToAccountID     *int      `json:"to_account_id,omitempty"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	Description     string    `json:"description"`
	CreatedAt       time.Time `json:"created_at"`
}

type Credit struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	AccountID      int       `json:"account_id"`
	Principal      float64   `json:"principal"`
	InterestRate   float64   `json:"interest_rate"`
	TermMonths     int       `json:"term_months"`
	MonthlyPayment float64   `json:"monthly_payment"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type PaymentSchedule struct {
	ID          int        `json:"id"`
	CreditID    int        `json:"credit_id"`
	PaymentDate time.Time  `json:"payment_date"`
	Amount      float64    `json:"amount"`
	Principal   float64    `json:"principal"`
	Interest    float64    `json:"interest"`
	Status      string     `json:"status"`
	PaidAt      *time.Time `json:"paid_at,omitempty"`
}

type CreateAccountRequest struct {
	Currency string `json:"currency"`
}

type DepositRequest struct {
	Amount float64 `json:"amount"`
}

func (r *DepositRequest) Validate() error {
	if r.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}

type TransferRequest struct {
	FromAccountID int     `json:"from_account_id"`
	ToAccountID   int     `json:"to_account_id"`
	Amount        float64 `json:"amount"`
}

func (r *TransferRequest) Validate() error {
	if r.FromAccountID == r.ToAccountID {
		return errors.New("source and destination accounts must differ")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be positive")
	}
	return nil
}

type CreateCreditRequest struct {
	AccountID  int     `json:"account_id"`
	Principal  float64 `json:"principal"`
	TermMonths int     `json:"term_months"`
}

func (r *CreateCreditRequest) Validate() error {
	if r.Principal <= 0 {
		return errors.New("principal must be positive")
	}
	if r.TermMonths < 1 || r.TermMonths > 360 {
		return errors.New("term must be between 1 and 360 months")
	}
	return nil
}

type Analytics struct {
	TotalIncome   float64 `json:"total_income"`
	TotalExpenses float64 `json:"total_expenses"`
	NetFlow       float64 `json:"net_flow"`
	Month         string  `json:"month"`
}

type CreditAnalytics struct {
	TotalDebt          float64 `json:"total_debt"`
	MonthlyObligations float64 `json:"monthly_obligations"`
	ActiveCredits      int     `json:"active_credits"`
}

type BalanceForecast struct {
	AccountID         int                `json:"account_id"`
	CurrentBalance    float64            `json:"current_balance"`
	ForecastDays      int                `json:"forecast_days"`
	PredictedBalance  float64            `json:"predicted_balance"`
	ScheduledPayments []ScheduledPayment `json:"scheduled_payments"`
}

type ScheduledPayment struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
	Type   string    `json:"type"`
}
