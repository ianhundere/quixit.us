package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"

	"sample-exchange/backend/config"
)

type EmailService struct {
	config *config.Config
	auth   smtp.Auth
}

func NewEmailService(cfg *config.Config) *EmailService {
	auth := smtp.PlainAuth("", cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPHost)
	return &EmailService{
		config: cfg,
		auth:   auth,
	}
}

func (s *EmailService) SendVerificationEmail(to, token string) error {
	subject := "Verify Your Email"
	verificationLink := fmt.Sprintf("http://localhost:%s/api/auth/verify/%s", s.config.Port, token)
	
	tmpl := `
	<h2>Welcome to Sample Exchange!</h2>
	<p>Please verify your email address by clicking the link below:</p>
	<p><a href="{{.Link}}">Verify Email</a></p>
	<p>If you didn't create this account, please ignore this email.</p>
	`
	
	t, err := template.New("verification").Parse(tmpl)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := t.Execute(&body, struct{ Link string }{Link: verificationLink}); err != nil {
		return err
	}

	msg := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", to, s.config.SMTPFrom, subject, body.String())

	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	return smtp.SendMail(addr, s.auth, s.config.SMTPFrom, []string{to}, []byte(msg))
} 