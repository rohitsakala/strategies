package utils

import (
	"crypto/tls"
	"os"

	gomail "gopkg.in/mail.v2"
)

func SendEmail(subject, body string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", "stratgies@rohitsakala.com")
	m.SetHeader("To", "rohitsakala@gmail.com")
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer("smtp.gmail.com", 587, "rohitsakala@gmail.com", os.Getenv("EMAIL_PASSWORD"))
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if err := d.DialAndSend(m); err != nil {
		return err
	}
}
