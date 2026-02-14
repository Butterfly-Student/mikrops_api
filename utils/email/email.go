package email

import (
	"fmt"
	"net/smtp"
	"strings"
)

type EmailUtil struct {
	smtpHost     string
	smtpPort     string
	smtpUser     string
	smtpPassword string
	fromEmail    string
	fromName     string
	isEnabled    bool
}

type EmailConfig struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	FromName     string
	IsEnabled    bool
}

func NewEmailUtil(config EmailConfig) *EmailUtil {
	return &EmailUtil{
		smtpHost:     config.SMTPHost,
		smtpPort:     config.SMTPPort,
		smtpUser:     config.SMTPUser,
		smtpPassword: config.SMTPPassword,
		fromEmail:    config.FromEmail,
		fromName:     config.FromName,
		isEnabled:    config.IsEnabled,
	}
}

func (e *EmailUtil) SendEmail(to string, subject string, content string) error {
	if !e.isEnabled {
		return fmt.Errorf("email sending is disabled")
	}

	// Create message
	message := fmt.Sprintf("From: %s <%s>\n", e.fromName, e.fromEmail)
	message += fmt.Sprintf("To: %s\n", to)
	message += fmt.Sprintf("Subject: %s\n", subject)
	message += "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	message += content

	// Send email
	auth := smtp.PlainAuth("", e.smtpUser, e.smtpPassword, e.smtpHost)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort),
		auth,
		e.fromEmail,
		[]string{to},
		[]byte(message),
	)

	return err
}

func (e *EmailUtil) SendEmailMultiple(to []string, subject string, content string) error {
	if !e.isEnabled {
		return fmt.Errorf("email sending is disabled")
	}

	for _, emailAddr := range to {
		if err := e.SendEmail(emailAddr, subject, content); err != nil {
			return err
		}
	}

	return nil
}

func (e *EmailUtil) SendHTMLEmail(to string, subject string, htmlContent string) error {
	if !e.isEnabled {
		return fmt.Errorf("email sending is disabled")
	}

	// Wrap HTML content in basic HTML structure
	content := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
	<meta charset="UTF-8">
	<title>%s</title>
	<style>
		body {
			font-family: Arial, sans-serif;
			line-height: 1.6;
			color: #333;
		}
		.container {
			max-width: 600px;
			margin: 0 auto;
			padding: 20px;
			background-color: #f9f9f9;
		}
		.content {
			background-color: white;
			padding: 30px;
			border-radius: 5px;
			box-shadow: 0 2px 4px rgba(0,0,0,0.1);
		}
		h1 {
			color: #2c3e50;
		}
		.footer {
			text-align: center;
			margin-top: 20px;
			font-size: 12px;
			color: #7f8c8d;
		}
	</style>
</head>
<body>
	<div class="container">
		<div class="content">
			%s
		</div>
		<div class="footer">
			<p>This is an automated email. Please do not reply.</p>
		</div>
	</div>
</body>
</html>`, subject, htmlContent)

	return e.SendEmail(to, subject, content)
}

func (e *EmailUtil) ValidateEmail(email string) bool {
	if email == "" {
		return false
	}

	// Basic email validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if parts[0] == "" || parts[1] == "" {
		return false
	}

	// Check for domain part
	domainParts := strings.Split(parts[1], ".")
	if len(domainParts) < 2 {
		return false
	}

	return true
}

func (e *EmailUtil) IsEnabled() bool {
	return e.isEnabled
}
