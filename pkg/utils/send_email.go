package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(to, otp string) error {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	user := os.Getenv("SMTP_USER")
	pass := os.Getenv("SMTP_PASSWORD")
	from := os.Getenv("SMTP_FROM")

	if from == "" {
		from = user
	}
	if host == "" || user == "" || pass == "" {
		return fmt.Errorf("SMTP config not set")
	}
	if port == "" {
		port = "587"
	}

	addr := host + ":" + port
	auth := smtp.PlainAuth("", user, pass, host)

	subject := "Your verification code"
	body := fmt.Sprintf("Your verification code is: %s", otp)

	msg := "From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
		"\r\n" + body

	if err := smtp.SendMail(addr, auth, from, []string{to}, []byte(msg)); err != nil {
		return fmt.Errorf("failed to send mail: %w", err)
	}
	return nil
}
