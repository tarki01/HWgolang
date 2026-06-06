package service

import (
	"errors"
	"time"

	models "HWGO/internal/Core"
	repository "HWGO/internal/Repositories"

	"github.com/sirupsen/logrus"
)

type AnalyticsService struct {
	accountRepo  *repository.AccountRepository
	txRepo       *repository.TransactionRepository
	creditRepo   *repository.CreditRepository
	scheduleRepo *repository.PaymentScheduleRepository
	log          *logrus.Logger
}

func NewAnalyticsService(
	accountRepo *repository.AccountRepository,
	txRepo *repository.TransactionRepository,
	creditRepo *repository.CreditRepository,
	scheduleRepo *repository.PaymentScheduleRepository,
	log *logrus.Logger,
) *AnalyticsService {
	return &AnalyticsService{
		accountRepo:  accountRepo,
		txRepo:       txRepo,
		creditRepo:   creditRepo,
		scheduleRepo: scheduleRepo,
		log:          log,
	}
}

func (s *AnalyticsService) GetMonthlyAnalytics(userID int) ([]models.Analytics, error) {
	accounts, err := s.accountRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var results []models.Analytics

	for m := 0; m < 6; m++ {
		month := now.AddDate(0, -m, 0)
		from := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0)

		var totalIncome, totalExpenses float64
		for _, acc := range accounts {
			inc, exp, err := s.txRepo.GetMonthlyStats(acc.ID, from, to)
			if err != nil {
				s.log.WithError(err).Warn("monthly stats error")
				continue
			}
			totalIncome += inc
			totalExpenses += exp
		}

		results = append(results, models.Analytics{
			TotalIncome:   totalIncome,
			TotalExpenses: totalExpenses,
			NetFlow:       totalIncome - totalExpenses,
			Month:         from.Format("2006-01"),
		})
	}
	return results, nil
}

func (s *AnalyticsService) GetCreditAnalytics(userID int) (*models.CreditAnalytics, error) {
	credits, err := s.creditRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	result := &models.CreditAnalytics{}
	for _, c := range credits {
		if c.Status != "active" {
			continue
		}
		result.ActiveCredits++
		result.MonthlyObligations += c.MonthlyPayment

		schedules, err := s.scheduleRepo.FindByCreditID(c.ID)
		if err != nil {
			continue
		}
		for _, ps := range schedules {
			if ps.Status == "pending" || ps.Status == "overdue" {
				result.TotalDebt += ps.Amount
			}
		}
	}
	return result, nil
}

func (s *AnalyticsService) GetBalanceForecast(userID, accountID, days int) (*models.BalanceForecast, error) {
	if days < 1 || days > 365 {
		return nil, errors.New("forecast period must be between 1 and 365 days")
	}

	acc, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("access denied")
	}

	until := time.Now().AddDate(0, 0, days)
	upcoming, err := s.scheduleRepo.GetUpcoming(accountID, until)
	if err != nil {
		return nil, err
	}

	predicted := acc.Balance
	payments := make([]models.ScheduledPayment, 0, len(upcoming))
	for _, sched := range upcoming {
		predicted -= sched.Amount
		payments = append(payments, models.ScheduledPayment{
			Date:   sched.PaymentDate,
			Amount: sched.Amount,
			Type:   "credit_payment",
		})
	}

	return &models.BalanceForecast{
		AccountID:         accountID,
		CurrentBalance:    acc.Balance,
		ForecastDays:      days,
		PredictedBalance:  predicted,
		ScheduledPayments: payments,
	}, nil
}
