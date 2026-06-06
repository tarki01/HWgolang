package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	config "HWGO/internal/Cfg"
	handler "HWGO/internal/Handlers"
	middleware "HWGO/internal/MidWare"
	migrations "HWGO/internal/Migrations"
	repository "HWGO/internal/Repositories"
	service "HWGO/internal/Services"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Load()

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	log.SetLevel(level)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.WithError(err).Fatal("db open failed")
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.WithError(err).Fatal("db ping failed")
	}
	log.Info("connected to database")

	if err := migrations.Run(db, log); err != nil {
		log.WithError(err).Fatal("migrations failed")
	}

	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	cardRepo := repository.NewCardRepository(db)
	txRepo := repository.NewTransactionRepository(db)
	creditRepo := repository.NewCreditRepository(db)
	scheduleRepo := repository.NewPaymentScheduleRepository(db)

	cbrSvc := service.NewCBRService(log)
	emailSvc := service.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, log)

	userSvc := service.NewUserService(userRepo, cfg.JWTSecret, log)
	accountSvc := service.NewAccountService(accountRepo, txRepo, emailSvc, userRepo, log)
	cardSvc := service.NewCardService(cardRepo, accountRepo, cfg.PGPKey, []byte(cfg.HMACSecret), log)
	creditSvc := service.NewCreditService(creditRepo, scheduleRepo, accountRepo, txRepo, userRepo, cbrSvc, emailSvc, log)
	analyticsSvc := service.NewAnalyticsService(accountRepo, txRepo, creditRepo, scheduleRepo, log)

	h := handler.New(userSvc, accountSvc, cardSvc, creditSvc, analyticsSvc, log)

	go runScheduler(creditSvc, log)

	r := mux.NewRouter()
	r.Use(loggingMiddleware(log))

	r.HandleFunc("/register", h.Register).Methods(http.MethodPost)
	r.HandleFunc("/login", h.Login).Methods(http.MethodPost)

	auth := r.PathPrefix("/").Subrouter()
	auth.Use(middleware.Auth(cfg.JWTSecret))

	auth.HandleFunc("/accounts", h.CreateAccount).Methods(http.MethodPost)
	auth.HandleFunc("/accounts", h.GetUserAccounts).Methods(http.MethodGet)
	auth.HandleFunc("/accounts/{accountId:[0-9]+}", h.GetAccount).Methods(http.MethodGet)
	auth.HandleFunc("/accounts/{accountId:[0-9]+}/deposit", h.Deposit).Methods(http.MethodPost)
	auth.HandleFunc("/accounts/{accountId:[0-9]+}/withdraw", h.Withdraw).Methods(http.MethodPost)
	auth.HandleFunc("/accounts/{accountId:[0-9]+}/cards", h.GetCards).Methods(http.MethodGet)
	auth.HandleFunc("/accounts/{accountId:[0-9]+}/predict", h.GetBalanceForecast).Methods(http.MethodGet)

	auth.HandleFunc("/cards", h.IssueCard).Methods(http.MethodPost)

	auth.HandleFunc("/transfer", h.Transfer).Methods(http.MethodPost)

	auth.HandleFunc("/credits", h.CreateCredit).Methods(http.MethodPost)
	auth.HandleFunc("/credits", h.GetUserCredits).Methods(http.MethodGet)
	auth.HandleFunc("/credits/{creditId:[0-9]+}/schedule", h.GetCreditSchedule).Methods(http.MethodGet)

	auth.HandleFunc("/analytics", h.GetAnalytics).Methods(http.MethodGet)

	addr := ":" + cfg.ServerPort
	log.WithField("addr", addr).Info("server starting")

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.WithError(err).Fatal("server stopped")
		os.Exit(1)
	}
}

func runScheduler(creditSvc *service.CreditService, log *logrus.Logger) {
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	log.Info("payment scheduler started")
	for range ticker.C {
		log.Info("running overdue payment check")
		creditSvc.ProcessOverduePayments()
	}
}

func loggingMiddleware(log *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.WithFields(logrus.Fields{
				"method":   r.Method,
				"path":     r.URL.Path,
				"duration": time.Since(start).String(),
			}).Info("request")
		})
	}
}
