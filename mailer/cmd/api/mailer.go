package main

import (
	"bytes"
	"github.com/vanng822/go-premailer/premailer"
	"html/template"
	"log/slog"
	"os"
	"strconv"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type Mailer struct {
	server *mail.SMTPServer
	domain string

	FromAddress string
	FromName    string
}

func NewMailer() (*Mailer, error) {
	port, err := strconv.Atoi(os.Getenv("MAIL_PORT"))
	if err != nil {
		return nil, err
	}

	srv := mail.NewSMTPClient()
	srv.Username = os.Getenv("MAIL_USERNAME")
	srv.Password = os.Getenv("MAIL_PASSWORD")
	srv.Host = os.Getenv("MAIL_HOST")
	srv.Port = port

	var enc mail.Encryption
	switch os.Getenv("MAIL_ENCRYPTION") {
	case "ssl":
		enc = mail.EncryptionSSLTLS
	case "none", "":
		enc = mail.EncryptionNone
	default:
		// Also the encryption for "tls" type
		enc = mail.EncryptionSTARTTLS
	}

	srv.Encryption = enc
	srv.KeepAlive = true
	srv.ConnectTimeout = 10 * time.Second

	return &Mailer{
		FromAddress: os.Getenv("MAIL_FROM_ADDRESS"),
		FromName:    os.Getenv("MAIL_FROM_NAME"),
		server:      srv,
		domain:      os.Getenv("MAIL_DOMAIN"),
	}, nil
}

type Message struct {
	Data    any
	DataMap map[string]any

	Attachments []string

	From     string
	FromName string
	To       string
	Subject  string
}

func (m *Mailer) SendMessage(email *mail.Email) error {
	client, err := m.server.Connect()
	if err != nil {
		return err
	}

	defer func() {
		err = client.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// Fallback if sending unsuccessful
	for i := 1; i <= 3; i++ {
		err = email.Send(client)

		if err == nil {
			return nil
		}

		if i == 3 {
			return err
		}

		<-ticker.C
	}

	return nil
}

func (m *Mailer) BuildMessage(msg Message) (*mail.Email, error) {
	// Sender Address
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	// Sender Name
	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	data := map[string]any{
		"content": msg.Data,
	}

	msg.DataMap = data

	plain, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return nil, err
	}

	formatted, err := m.buildHTMLMessage(msg)
	if err != nil {
		return nil, err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject).
		SetBody(mail.TextPlain, plain).
		AddAlternative(mail.TextHTML, formatted)

	for _, at := range msg.Attachments {
		email.AddAttachment(at)
	}

	return email, nil
}

func (m *Mailer) buildHTMLMessage(msg Message) (string, error) {
	ts, err := m.buildTemplate("./templates/mail.html.gohtml", "email", msg.DataMap)
	if err != nil {
		return "", err
	}

	// get the formatted (html) message from the template string
	fm, err := m.inlineCSS(ts)
	if err != nil {
		return "", err
	}

	return fm, nil
}

func (m *Mailer) buildPlainTextMessage(msg Message) (string, error) {
	return m.buildTemplate("./templates/mail.plain.gohtml", "email-plain", msg.DataMap)
}

func (m *Mailer) buildTemplate(templateToRender, name string, data any) (string, error) {
	t, err := template.New(name).ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}

func (m *Mailer) inlineCSS(s string) (string, error) {
	opts := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &opts)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
