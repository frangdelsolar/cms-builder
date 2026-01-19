package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailSender struct {
	Host     string
	Port     string
	User     string
	Password string
	Sender   string
	FromName string
}

func NewEmailSender(host, port, user, password, sender, fromName string) *EmailSender {
	return &EmailSender{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Sender:   sender,
		FromName: fromName,
	}
}

func (s *EmailSender) SendEmail(to []string, subject, body string, fromNameOverride string) error {
	if len(to) == 0 {
		return fmt.Errorf("recipient list cannot be empty")
	}

	displayName := s.FromName
	if fromNameOverride != "" {
		displayName = fromNameOverride
	}

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	toHeader := strings.Join(to, ",")

	fromHeader := fmt.Sprintf("%s <%s>", displayName, s.Sender)

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-version: 1.0;\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
			"\r\n"+
			"%s\r\n",
		fromHeader,
		toHeader,
		subject,
		body,
	))

	err := smtp.SendMail(addr, auth, s.Sender, to, msg)
	if err != nil {
		return fmt.Errorf("failed to send email to %v: %w", to, err)
	}
	return nil
}
