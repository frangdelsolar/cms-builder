package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailSender handles email sending configuration and operations.
type EmailSender struct {
	Host     string
	Port     string
	User     string
	Password string
	Sender   string
}

// NewEmailSender creates a new email sender instance.
func NewEmailSender(host, port, user, password, sender string) *EmailSender {
	return &EmailSender{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Sender:   sender,
	}
}

// SendEmail sends an HTML email to the specified recipients.
// Returns an error if the recipient list is empty or if sending fails.
func (s *EmailSender) SendEmail(to []string, subject, body string) error {
	if len(to) == 0 {
		return fmt.Errorf("recipient list cannot be empty")
	}

	auth := smtp.PlainAuth("", s.User, s.Password, s.Host)
	addr := fmt.Sprintf("%s:%s", s.Host, s.Port)

	toHeader := strings.Join(to, ",")

	msg := []byte(fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-version: 1.0;\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
			"\r\n"+
			"%s\r\n",
		s.Sender,
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
