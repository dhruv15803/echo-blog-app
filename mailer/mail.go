package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"os"

	"gopkg.in/gomail.v2"
)

type InviteMailData struct {
	Subject       string
	ActivationUrl string
}

type PasswordResetMailData struct {
	Subject           string
	PasswordResetLink string
}

type GoMailConfig struct {
	GoMailUsername string
	GoMailPassword string
	GoMailPort     int
}

func NewGoMailConfig(goMailUsername string, goMailPassword string, goMailPort int) *GoMailConfig {
	return &GoMailConfig{
		GoMailUsername: goMailUsername,
		GoMailPassword: goMailPassword,
		GoMailPort:     goMailPort,
	}
}

func SendGoInvitationMail(fromEmail string, toEmail string, subject string, templatePath string, plainTextToken string) error {
	goMailCfg := NewGoMailConfig(os.Getenv("GOMAIL_USERNAME"), os.Getenv("GOMAIL_PASSWORD"), 587)

	clientUrl := os.Getenv("CLIENT_URL")
	activationUrl := fmt.Sprintf("%s/activate-account/%s", clientUrl, plainTextToken)

	// parse template
	tmpl := template.Must(template.ParseFiles(templatePath))

	var body bytes.Buffer

	tmpl.Execute(&body, InviteMailData{Subject: subject, ActivationUrl: activationUrl})

	message := gomail.NewMessage()

	message.SetHeader("From", fromEmail)
	message.SetHeader("To", toEmail)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body.String())

	dialer := gomail.NewDialer("smtp.gmail.com", 587, goMailCfg.GoMailUsername, goMailCfg.GoMailPassword)

	return dialer.DialAndSend(message)
}

func SendGoPasswordResetMail(fromEmail string, toEmail string, subject string, templatePath string, plainTextToken string) error {
	goMailCfg := NewGoMailConfig(os.Getenv("GOMAIL_USERNAME"), os.Getenv("GOMAIL_PASSWORD"), 587)
	clientUrl := os.Getenv("CLIENT_URL")

	passwordResetLink := fmt.Sprintf("%s/password-reset/%s", clientUrl, plainTextToken)

	tmpl := template.Must(template.ParseFiles(templatePath))

	var body bytes.Buffer

	if err := tmpl.Execute(&body, PasswordResetMailData{Subject: "Password reset mail", PasswordResetLink: passwordResetLink}); err != nil {
		return err
	}

	message := gomail.NewMessage()

	message.SetHeader("From", fromEmail)
	message.SetHeader("To", toEmail)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body.String())

	dialer := gomail.NewDialer("smtp.gmail.com", goMailCfg.GoMailPort, goMailCfg.GoMailUsername, goMailCfg.GoMailPassword)

	return dialer.DialAndSend(message)
}
