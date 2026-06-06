package service

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"github.com/sirupsen/logrus"
)

type EmailService struct {
	host string
	port int
	user string
	pass string
	log  *logrus.Logger
}

func NewEmailService(host string, port int, user, pass string, log *logrus.Logger) *EmailService {
	return &EmailService{host: host, port: port, user: user, pass: pass, log: log}
}

func (s *EmailService) SendPaymentNotification(to string, amount float64, creditID int) error {
	subject := "Списание платежа по кредиту"
	body := fmt.Sprintf("Списание по кредиту #%d на сумму %.2f RUB выполнено успешно.", creditID, amount)
	return s.send(to, subject, body)
}

func (s *EmailService) SendOverdueNotification(to string, amount float64, creditID int) error {
	subject := "Просроченный платёж по кредиту"
	body := fmt.Sprintf("По кредиту #%d начислен штраф за просрочку. Итоговая сумма: %.2f RUB.", creditID, amount)
	return s.send(to, subject, body)
}

func (s *EmailService) SendTransferNotification(to string, amount float64, fromID, toID int) error {
	subject := "Перевод выполнен"
	body := fmt.Sprintf("Перевод %.2f RUB со счёта #%d на счёт #%d выполнен.", amount, fromID, toID)
	return s.send(to, subject, body)
}

func (s *EmailService) send(to, subject, body string) error {
	if s.user == "" {
		s.log.Debug("SMTP not configured, skipping email")
		return nil
	}

	addr := fmt.Sprintf("%s:%d", s.host, s.port)
	msg := buildMessage(s.user, to, subject, body)

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		s.log.WithError(err).Error("smtp connect failed")
		return fmt.Errorf("smtp connect: %w", err)
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	// попробуем STARTTLS, но не падаем если сервер не поддерживает
	tlsCfg := &tls.Config{ServerName: s.host, InsecureSkipVerify: true}
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsCfg); err != nil {
			s.log.WithError(err).Warn("STARTTLS failed, continuing plain")
		}
	}

	// аутентификация опциональна
	if s.pass != "" {
		auth := smtp.PlainAuth("", s.user, s.pass, s.host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	if err := client.Mail(s.user); err != nil {
		return fmt.Errorf("smtp MAIL: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("smtp RCPT: %w", err)
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp DATA: %w", err)
	}
	defer w.Close()
	if _, err := fmt.Fprint(w, msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	s.log.WithField("to", to).Info("email sent")
	return nil
}

func buildMessage(from, to, subject, body string) string {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + subject + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return b.String()
}
