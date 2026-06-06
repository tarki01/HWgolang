package service

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	models "HWGO/internal/Core"
	repository "HWGO/internal/Repositories"
	"HWGO/pkg/crypto"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	repo      *repository.UserRepository
	jwtSecret string
	log       *logrus.Logger
}

func NewUserService(repo *repository.UserRepository, jwtSecret string, log *logrus.Logger) *UserService {
	return &UserService{repo: repo, jwtSecret: jwtSecret, log: log}
}

func (s *UserService) Register(req *models.RegisterRequest) (*models.User, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	emailExists, err := s.repo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	if emailExists {
		return nil, errors.New("email already taken")
	}

	usernameExists, err := s.repo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	if usernameExists {
		return nil, errors.New("username already taken")
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("password hash: %w", err)
	}

	user, err := s.repo.Create(req.Username, req.Email, hash)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	s.log.WithField("user_id", user.ID).Info("user registered")
	return user, nil
}

func (s *UserService) Login(req *models.LoginRequest) (string, *models.User, error) {
	if err := req.Validate(); err != nil {
		return "", nil, err
	}

	user, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	if err := crypto.CheckPassword(user.PasswordHash, req.Password); err != nil {
		return "", nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", nil, fmt.Errorf("token generation: %w", err)
	}

	s.log.WithField("user_id", user.ID).Info("user logged in")
	return token, user, nil
}

func (s *UserService) GetByID(id int) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) generateToken(userID int) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   strconv.Itoa(userID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
