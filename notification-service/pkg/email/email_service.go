package email

import (
	"context"
	"fmt"
	"html/template"
	"micro-warehouse/notificaiton-service/configs"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"gopkg.in/gomail.v2"
)

type EmailServiceInterface interface {
	SendWelcomeEmail(ctx context.Context, payload EmailPayload) error
	SendCustomEmail(ctx context.Context, to, subject, body string) error
}

type EmailPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Type     string `json:"type"`
	UserID   uint   `json:"user_id"`
	Name     string `json:"name"`
}

type emailService struct {
	cfg configs.Config
}

// SendCustomEmail implements EmailServiceInterface.
func (e *emailService) SendCustomEmail(ctx context.Context, to string, subject string, body string) error {
	if e.cfg.Email.Host == "" || e.cfg.Email.User == "" || e.cfg.Email.Password == "" {
		log.Errorf("[EmailService] SendCustomEmail - 1: %v", "email configuration is incomplete")
		return fmt.Errorf("email configuration is incomplete: Host=%s, User=%s", e.cfg.Email.Host, e.cfg.Email.User)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", e.cfg.Email.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.cfg.Email.Host, e.cfg.Email.Port, e.cfg.Email.User, e.cfg.Email.Password)

	if err := d.DialAndSend(m); err != nil {
		log.Errorf("[EmailService] SendCustomEmail - 2: %v", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// SendWelcomeEmail implements EmailServiceInterface.
func (e *emailService) SendWelcomeEmail(ctx context.Context, payload EmailPayload) error {
	// Validasi konfigurasi email
	if e.cfg.Email.Host == "" || e.cfg.Email.User == "" || e.cfg.Email.Password == "" {
		log.Errorf("[EmailService] SendWelcomeEmail - Email configuration is incomplete: Host=%s, User=%s", e.cfg.Email.Host, e.cfg.Email.User)
		return fmt.Errorf("email configuration is incomplete")
	}

	subject := "Selamat Datang di Warehouse Management System"

	htmlTemplate := `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>Selamat Datang</title>
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
				.button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; }
			</style>
		</head>
		<body>
			<div class="container">
				<div class="header">
					<h1>Selamat Datang!</h1>
				</div>
				<div class="content">
					<h2>Halo {{.Name}},</h2>
					<p>Selamat datang di Warehouse Management System! Akun Anda telah berhasil dibuat.</p>
					<p><strong>Email:</strong> {{.Email}}</p>
					<p><strong>Password:</strong> {{.Password}}</p>
					<p>Silakan login dengan kredensial di atas dan jangan lupa untuk mengganti password Anda setelah login pertama kali.</p>
					<br>
					<p>Terima kasih telah bergabung dengan kami!</p>
				</div>
				<div class="footer">
					<p>Email ini dikirim otomatis, mohon tidak membalas email ini.</p>
				</div>
			</div>
		</body>
		</html>`
	tmpl, err := template.New("welcome").Parse(htmlTemplate)
	if err != nil {
		log.Errorf("[EmailService] SendWelcomeEmail - 1: %v", err)
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var body strings.Builder
	err = tmpl.Execute(&body, payload)
	if err != nil {
		log.Errorf("[EmailService] SendWelcomeEmail - 2: %v", err)
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	err = e.SendCustomEmail(ctx, payload.Email, subject, body.String())
	if err != nil {
		log.Errorf("[EmailService] SendWelcomeEmail - 3: %v", err)
		return fmt.Errorf("failed to send welcome email: %v", err)
	}

	return nil
}

func NewEmailService(cfg configs.Config) EmailServiceInterface {
	return &emailService{
		cfg: cfg,
	}
}
