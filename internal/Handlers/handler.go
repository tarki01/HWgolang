package handler

import (
	"encoding/json"
	"net/http"

	service "HWGO/internal/Services"

	"github.com/sirupsen/logrus"
)

type Handler struct {
	userSvc      *service.UserService
	accountSvc   *service.AccountService
	cardSvc      *service.CardService
	creditSvc    *service.CreditService
	analyticsSvc *service.AnalyticsService
	log          *logrus.Logger
}

func New(
	userSvc *service.UserService,
	accountSvc *service.AccountService,
	cardSvc *service.CardService,
	creditSvc *service.CreditService,
	analyticsSvc *service.AnalyticsService,
	log *logrus.Logger,
) *Handler {
	return &Handler{
		userSvc:      userSvc,
		accountSvc:   accountSvc,
		cardSvc:      cardSvc,
		creditSvc:    creditSvc,
		analyticsSvc: analyticsSvc,
		log:          log,
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
