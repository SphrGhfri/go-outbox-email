package email

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"outbox/config"
	"path/filepath"
	"strings"
)

// 1) Interface definition
type IEmailService interface {
	SendEmail(to []string, subject string, templateName string, data interface{}) error
}

// 2) Struct with SMTP auth + config
type EmailService struct {
	cfg         *config.Config
	auth        smtp.Auth
	templateDir string
}

// 3) Constructor
func NewEmailService(cfg *config.Config) IEmailService {
	auth := smtp.PlainAuth(
		"",
		cfg.EmailSMTPUsername,
		cfg.EmailSMTPPassword,
		cfg.EmailSMTPHost,
	)
	return &EmailService{
		cfg:         cfg,
		auth:        auth,
		templateDir: "./template/", // Adjust if your template is stored elsewhere
	}
}

func (s *EmailService) SendEmail(to []string, subject string, templateName string, data interface{}) error {
	// Parse the HTML template
	tmplPath := filepath.Join(s.templateDir, templateName)
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Render the template with data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %v", err)
	}

	// Prepare email headers
	headers := make(map[string]string)
	headers["From"] = s.cfg.EmailSender
	headers["To"] = strings.Join(to, ",")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Build the email message
	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body.String())

	// Connect to the SMTP server
	smtpAddr := fmt.Sprintf("%s:%d", s.cfg.EmailSMTPHost, s.cfg.EmailSMTPPort)

	// Send the email
	err = smtp.SendMail(
		smtpAddr,
		s.auth,
		s.cfg.EmailSender,
		to,
		[]byte(message.String()),
	)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
