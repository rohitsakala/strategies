package utils

import (
	"crypto/tls"
	"os"

	gomail "gopkg.in/mail.v2"
)

func SendEmail(subject, body string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", os.Getenv("SENDER_EMAIL_ADDRESS"))
	m.SetHeader("To", os.Getenv("EMAIL_ADDRESS"))
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, os.Getenv("SENDER_EMAIL_ADDRESS"), os.Getenv("SENDER_EMAIL_PASSWORD"))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	return nil
}
