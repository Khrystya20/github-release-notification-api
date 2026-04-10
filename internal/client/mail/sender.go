package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gopkg.in/gomail.v2"

	"github-release-notification-api/internal/config"
)

type Sender struct {
	host     string
	port     int
	username string
	password string
	from     string
	baseURL  string
}

func NewSender(cfg *config.Config) *Sender {
	port := 587
	if cfg.SMTPPort != "" {
		if parsedPort, err := strconv.Atoi(cfg.SMTPPort); err == nil {
			port = parsedPort
		} else {
			log.Printf("invalid SMTP_PORT %q, using default port %d", cfg.SMTPPort, port)
		}
	}

	return &Sender{
		host:     cfg.SMTPHost,
		port:     port,
		username: cfg.SMTPUsername,
		password: cfg.SMTPPassword,
		from:     cfg.SMTPFrom,
		baseURL:  strings.TrimRight(cfg.AppBaseURL, "/"),
	}
}

func (s *Sender) SendConfirmationEmail(email, confirmToken, unsubscribeToken, repo string) error {
	confirmURL := fmt.Sprintf("%s/api/confirm/%s", s.baseURL, confirmToken)
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", s.baseURL, unsubscribeToken)

	subject := fmt.Sprintf("Confirm subscription for %s", repo)
	body := fmt.Sprintf(
		"Hello!\n\n"+
			"You subscribed to notifications for repository: %s\n\n"+
			"Please confirm your subscription:\n%s\n\n"+
			"If this was not you, unsubscribe here:\n%s\n",
		repo,
		confirmURL,
		unsubscribeURL,
	)

	return s.send(email, subject, body)
}

func (s *Sender) SendNewReleaseEmail(email, repo, tag, unsubscribeToken string) error {
	unsubscribeURL := fmt.Sprintf("%s/api/unsubscribe/%s", s.baseURL, unsubscribeToken)

	subject := fmt.Sprintf("New release in %s: %s", repo, tag)
	body := fmt.Sprintf(
		"Hello!\n\n"+
			"A new release was detected.\n\n"+
			"Repository: %s\n"+
			"Tag: %s\n\n"+
			"To unsubscribe from these notifications, use this link:\n%s\n",
		repo,
		tag,
		unsubscribeURL,
	)

	return s.send(email, subject, body)
}

func (s *Sender) send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if s.username != "" || s.password != "" {
		d.TLSConfig = &tls.Config{
			ServerName: s.host,
			MinVersion: tls.VersionTLS12,
		}
	}

	if err := d.DialAndSend(m); err != nil {
		log.Printf("failed to send email to %s, subject=%q: %v", to, subject, err)
		return err
	}

	log.Printf("email sent successfully to %s, subject=%q", to, subject)
	return nil
}
