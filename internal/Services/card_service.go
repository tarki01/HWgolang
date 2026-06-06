package service

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	models "HWGO/internal/Core"
	repository "HWGO/internal/Repositories"
	"HWGO/pkg/crypto"

	"github.com/sirupsen/logrus"
)

type CardService struct {
	cardRepo    *repository.CardRepository
	accountRepo *repository.AccountRepository
	pgpKey      string
	hmacSecret  []byte
	log         *logrus.Logger
}

func NewCardService(
	cardRepo *repository.CardRepository,
	accountRepo *repository.AccountRepository,
	pgpKey string,
	hmacSecret []byte,
	log *logrus.Logger,
) *CardService {
	return &CardService{
		cardRepo:    cardRepo,
		accountRepo: accountRepo,
		pgpKey:      pgpKey,
		hmacSecret:  hmacSecret,
		log:         log,
	}
}

func (s *CardService) IssueCard(userID, accountID int) (*models.CardView, error) {
	acc, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("access denied")
	}

	number := generateLuhnNumber()
	expiry := generateExpiry()
	cvv := generateCVV()

	encNumber, err := crypto.EncryptPGP(number, s.pgpKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt card number: %w", err)
	}
	encExpiry, err := crypto.EncryptPGP(expiry, s.pgpKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt expiry: %w", err)
	}
	cvvHash, err := crypto.HashCVV(cvv)
	if err != nil {
		return nil, fmt.Errorf("hash cvv: %w", err)
	}
	hmacNum := crypto.ComputeHMAC(number, s.hmacSecret)
	lastFour := number[len(number)-4:]

	card, err := s.cardRepo.Create(accountID, encNumber, encExpiry, cvvHash, hmacNum, lastFour)
	if err != nil {
		return nil, fmt.Errorf("save card: %w", err)
	}

	s.log.WithFields(logrus.Fields{"card_id": card.ID, "account_id": accountID}).Info("card issued")

	return &models.CardView{
		ID:        card.ID,
		AccountID: card.AccountID,
		Number:    number,
		Expiry:    expiry,
		LastFour:  lastFour,
	}, nil
}

func (s *CardService) GetCards(userID, accountID int) ([]models.CardView, error) {
	acc, err := s.accountRepo.FindByID(accountID)
	if err != nil {
		return nil, err
	}
	if acc.UserID != userID {
		return nil, errors.New("access denied")
	}

	cards, err := s.cardRepo.FindByAccountID(accountID)
	if err != nil {
		return nil, err
	}

	views := make([]models.CardView, 0, len(cards))
	for _, c := range cards {
		number, err := crypto.DecryptPGP(c.EncryptedNumber, s.pgpKey)
		if err != nil {
			s.log.WithError(err).WithField("card_id", c.ID).Error("decrypt card number failed")
			number = "****" + c.LastFour
		}
		expiry, err := crypto.DecryptPGP(c.EncryptedExpiry, s.pgpKey)
		if err != nil {
			expiry = "**/**"
		}
		views = append(views, models.CardView{
			ID:        c.ID,
			AccountID: c.AccountID,
			Number:    number,
			Expiry:    expiry,
			LastFour:  c.LastFour,
		})
	}
	return views, nil
}

func generateLuhnNumber() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	digits := make([]int, 15)
	digits[0] = r.Intn(9) + 1
	for i := 1; i < 15; i++ {
		digits[i] = r.Intn(10)
	}

	sum := 0
	for i, d := range digits {
		if (15-i)%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	check := (10 - (sum % 10)) % 10

	result := ""
	for _, d := range digits {
		result += strconv.Itoa(d)
	}
	result += strconv.Itoa(check)
	return result
}

func generateExpiry() string {
	now := time.Now()
	exp := now.AddDate(3, 0, 0)
	return fmt.Sprintf("%02d/%02d", exp.Month(), exp.Year()%100)
}

func generateCVV() string {
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	return fmt.Sprintf("%03d", r.Intn(1000))
}
