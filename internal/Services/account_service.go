package service

import (
	"errors"
	"fmt"

	models "HWGO/internal/Core"
	repository "HWGO/internal/Repositories"

	"github.com/sirupsen/logrus"
)

type AccountService struct {
	accountRepo *repository.AccountRepository
	txRepo      *repository.TransactionRepository
	emailSvc    *EmailService
	userRepo    *repository.UserRepository
	log         *logrus.Logger
}

func NewAccountService(
	accountRepo *repository.AccountRepository,
	txRepo *repository.TransactionRepository,
	emailSvc *EmailService,
	userRepo *repository.UserRepository,
	log *logrus.Logger,
) *AccountService {
	return &AccountService{
		accountRepo: accountRepo,
		txRepo:      txRepo,
		emailSvc:    emailSvc,
		userRepo:    userRepo,
		log:         log,
	}
}

func (s *AccountService) CreateAccount(userID int, req *models.CreateAccountRequest) (*models.Account, error) {
	currency := req.Currency
	if currency == "" {
		currency = "RUB"
	}
	if currency != "RUB" {
		return nil, errors.New("only RUB accounts are supported")
	}

	acc, err := s.accountRepo.Create(userID, currency)
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}
	s.log.WithFields(logrus.Fields{"user_id": userID, "account_id": acc.ID}).Info("account created")
	return acc, nil
}

func (s *AccountService) GetUserAccounts(userID int) ([]models.Account, error) {
	return s.accountRepo.FindByUserID(userID)
}

func (s *AccountService) GetAccount(userID, accountID int) (*models.Account, error) {
	acc, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("access denied")
	}
	return acc, nil
}

func (s *AccountService) Deposit(userID, accountID int, req *models.DepositRequest) (*models.Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	acc, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("access denied")
	}

	newBalance := acc.Balance + req.Amount
	if err := s.accountRepo.UpdateBalance(nil, accountID, newBalance); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	toID := accountID
	if _, err := s.txRepo.Create(nil, nil, &toID, req.Amount, "deposit", "deposit"); err != nil {
		s.log.WithError(err).Warn("failed to record deposit transaction")
	}

	acc.Balance = newBalance
	s.log.WithFields(logrus.Fields{"account_id": accountID, "amount": req.Amount}).Info("deposit completed")
	return acc, nil
}

func (s *AccountService) Withdraw(userID, accountID int, req *models.DepositRequest) (*models.Account, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	acc, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("access denied")
	}
	if acc.Balance < req.Amount {
		return nil, errors.New("insufficient funds")
	}

	newBalance := acc.Balance - req.Amount
	if err := s.accountRepo.UpdateBalance(nil, accountID, newBalance); err != nil {
		return nil, fmt.Errorf("update balance: %w", err)
	}

	fromID := accountID
	if _, err := s.txRepo.Create(nil, &fromID, nil, req.Amount, "withdrawal", "withdrawal"); err != nil {
		s.log.WithError(err).Warn("failed to record withdrawal transaction")
	}

	acc.Balance = newBalance
	return acc, nil
}

func (s *AccountService) Transfer(userID int, req *models.TransferRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}

	tx, err := s.accountRepo.BeginTx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	from, err := s.accountRepo.FindByIDTx(tx, req.FromAccountID)
	if err != nil {
		return err
	}
	if from.UserID != userID {
		return errors.New("access denied")
	}
	if from.Balance < req.Amount {
		return errors.New("insufficient funds")
	}

	to, err := s.accountRepo.FindByIDTx(tx, req.ToAccountID)
	if err != nil {
		return err
	}

	if err := s.accountRepo.UpdateBalance(tx, from.ID, from.Balance-req.Amount); err != nil {
		return fmt.Errorf("debit: %w", err)
	}
	if err := s.accountRepo.UpdateBalance(tx, to.ID, to.Balance+req.Amount); err != nil {
		return fmt.Errorf("credit: %w", err)
	}

	fromID, toID := from.ID, to.ID
	if _, err := s.txRepo.Create(tx, &fromID, &toID, req.Amount, "transfer", "transfer"); err != nil {
		return fmt.Errorf("record tx: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	s.log.WithFields(logrus.Fields{
		"from": from.ID, "to": to.ID, "amount": req.Amount,
	}).Info("transfer completed")

	user, err := s.userRepo.FindByID(userID)
	if err == nil {
		go s.emailSvc.SendTransferNotification(user.Email, req.Amount, from.ID, to.ID)
	}

	return nil
}
