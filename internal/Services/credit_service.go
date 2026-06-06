package service

import (
	"fmt"
	"math"
	"time"

	models "HWGO/internal/Core"
	repository "HWGO/internal/Repositories"

	"github.com/sirupsen/logrus"
)

type CreditService struct {
	creditRepo   *repository.CreditRepository
	scheduleRepo *repository.PaymentScheduleRepository
	accountRepo  *repository.AccountRepository
	txRepo       *repository.TransactionRepository
	userRepo     *repository.UserRepository
	cbrSvc       *CBRService
	emailSvc     *EmailService
	log          *logrus.Logger
}

func NewCreditService(
	creditRepo *repository.CreditRepository,
	scheduleRepo *repository.PaymentScheduleRepository,
	accountRepo *repository.AccountRepository,
	txRepo *repository.TransactionRepository,
	userRepo *repository.UserRepository,
	cbrSvc *CBRService,
	emailSvc *EmailService,
	log *logrus.Logger,
) *CreditService {
	return &CreditService{
		creditRepo:   creditRepo,
		scheduleRepo: scheduleRepo,
		accountRepo:  accountRepo,
		txRepo:       txRepo,
		userRepo:     userRepo,
		cbrSvc:       cbrSvc,
		emailSvc:     emailSvc,
		log:          log,
	}
}

func (s *CreditService) CreateCredit(userID int, req *models.CreateCreditRequest) (*models.Credit, []models.PaymentSchedule, error) {
	if err := req.Validate(); err != nil {
		return nil, nil, err
	}

	acc, err := s.accountRepo.FindByID(req.AccountID)
	if err != nil {
		return nil, nil, err
	}
	if acc.UserID != userID {
		return nil, nil, fmt.Errorf("access denied")
	}

	rate, err := s.cbrSvc.GetKeyRate()
	if err != nil {
		s.log.WithError(err).Warn("CBR unavailable, using fallback rate")
		rate = 21.0
	}

	monthly := calculateAnnuity(req.Principal, rate, req.TermMonths)

	tx, err := s.creditRepo.GetDB().Begin()
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	credit, err := s.creditRepo.Create(tx, userID, req.AccountID, req.Principal, rate, req.TermMonths, monthly)
	if err != nil {
		return nil, nil, fmt.Errorf("create credit: %w", err)
	}

	schedules := buildSchedule(credit)
	if err := s.scheduleRepo.CreateBatch(tx, schedules); err != nil {
		return nil, nil, fmt.Errorf("create schedule: %w", err)
	}

	newBalance := acc.Balance + req.Principal
	if err := s.accountRepo.UpdateBalance(tx, req.AccountID, newBalance); err != nil {
		return nil, nil, fmt.Errorf("credit disbursement: %w", err)
	}

	toID := req.AccountID
	if _, err := s.txRepo.Create(tx, nil, &toID, req.Principal, "credit_disbursement", fmt.Sprintf("credit #%d disbursement", credit.ID)); err != nil {
		return nil, nil, fmt.Errorf("record tx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, nil, fmt.Errorf("commit: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"credit_id": credit.ID, "user_id": userID, "amount": req.Principal,
	}).Info("credit created")

	return credit, schedules, nil
}

func (s *CreditService) GetSchedule(userID, creditID int) ([]models.PaymentSchedule, error) {
	credit, err := s.creditRepo.FindByID(creditID)
	if err != nil {
		return nil, err
	}
	if credit.UserID != userID {
		return nil, fmt.Errorf("access denied")
	}
	return s.scheduleRepo.FindByCreditID(creditID)
}

func (s *CreditService) GetUserCredits(userID int) ([]models.Credit, error) {
	return s.creditRepo.FindByUserID(userID)
}

func (s *CreditService) ProcessOverduePayments() {
	schedules, err := s.creditRepo.GetOverdueSchedules()
	if err != nil {
		s.log.WithError(err).Error("failed to fetch overdue schedules")
		return
	}

	for _, sched := range schedules {
		s.processScheduledPayment(sched)
	}
}

func (s *CreditService) processScheduledPayment(sched models.PaymentSchedule) {
	credit, err := s.creditRepo.FindByID(sched.CreditID)
	if err != nil {
		s.log.WithError(err).WithField("schedule_id", sched.ID).Error("credit lookup failed")
		return
	}

	tx, err := s.creditRepo.GetDB().Begin()
	if err != nil {
		s.log.WithError(err).Error("begin tx failed")
		return
	}
	defer tx.Rollback()

	acc, err := s.accountRepo.FindByIDTx(tx, credit.AccountID)
	if err != nil {
		s.log.WithError(err).Error("account lookup failed")
		return
	}

	amount := sched.Amount
	if acc.Balance < amount {
		penaltyAmount := math.Round(amount*1.1*100) / 100
		if err := s.scheduleRepo.UpdateAmount(sched.ID, penaltyAmount); err != nil {
			s.log.WithError(err).Error("penalty update failed")
		}
		if err := s.scheduleRepo.MarkOverdue(sched.ID); err != nil {
			s.log.WithError(err).Error("mark overdue failed")
		}
		tx.Rollback()

		user, _ := s.userRepo.FindByID(credit.UserID)
		if user != nil {
			go s.emailSvc.SendOverdueNotification(user.Email, penaltyAmount, credit.ID)
		}
		s.log.WithField("schedule_id", sched.ID).Warn("insufficient funds, penalty applied")
		return
	}

	if err := s.accountRepo.UpdateBalance(tx, acc.ID, acc.Balance-amount); err != nil {
		s.log.WithError(err).Error("balance update failed")
		return
	}

	fromID := acc.ID
	if _, err := s.txRepo.Create(tx, &fromID, nil, amount, "credit_payment",
		fmt.Sprintf("credit #%d payment", credit.ID)); err != nil {
		s.log.WithError(err).Error("transaction record failed")
		return
	}

	if err := s.scheduleRepo.MarkPaid(tx, sched.ID); err != nil {
		s.log.WithError(err).Error("mark paid failed")
		return
	}

	if err := tx.Commit(); err != nil {
		s.log.WithError(err).Error("commit failed")
		return
	}

	user, _ := s.userRepo.FindByID(credit.UserID)
	if user != nil {
		go s.emailSvc.SendPaymentNotification(user.Email, amount, credit.ID)
	}

	s.log.WithFields(logrus.Fields{
		"schedule_id": sched.ID, "credit_id": credit.ID, "amount": amount,
	}).Info("payment processed")
}

func calculateAnnuity(principal, annualRate float64, months int) float64 {
	monthlyRate := annualRate / 100 / 12
	if monthlyRate == 0 {
		return math.Round(principal/float64(months)*100) / 100
	}
	factor := math.Pow(1+monthlyRate, float64(months))
	payment := principal * (monthlyRate * factor) / (factor - 1)
	return math.Round(payment*100) / 100
}

func buildSchedule(credit *models.Credit) []models.PaymentSchedule {
	schedules := make([]models.PaymentSchedule, 0, credit.TermMonths)

	balance := credit.Principal
	monthlyRate := credit.InterestRate / 100 / 12

	startDate := time.Now().AddDate(0, 1, 0)
	startDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, time.UTC)

	for i := 0; i < credit.TermMonths; i++ {
		interest := math.Round(balance*monthlyRate*100) / 100
		principal := math.Round((credit.MonthlyPayment-interest)*100) / 100
		if i == credit.TermMonths-1 {
			principal = math.Round(balance*100) / 100
		}
		if principal < 0 {
			principal = 0
		}
		amount := math.Round((principal+interest)*100) / 100

		schedules = append(schedules, models.PaymentSchedule{
			CreditID:    credit.ID,
			PaymentDate: startDate.AddDate(0, i, 0),
			Amount:      amount,
			Principal:   principal,
			Interest:    interest,
		})

		balance -= principal
		if balance < 0 {
			balance = 0
		}
	}
	return schedules
}
