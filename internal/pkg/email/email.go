package email

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"net/smtp"

	"gitlab.com/ayaka/config"
)

type MailSMTP struct {
	Conf  *config.Config `inject:"config"`
	Email smtp.Auth
	Host  string
	Port  string
}

func (m *MailSMTP) Startup() error {
	smtpHost := m.Conf.Email.SMTPHost
	smtpPort := m.Conf.Email.SMTPPort
	auth := smtp.PlainAuth("", m.Conf.Email.User, m.Conf.Email.Pass, smtpHost)

	m.Email = auth
	m.Host = smtpHost
	m.Port = smtpPort
	return nil
}

func (m *MailSMTP) Shutdown() error {
	return nil
}

func (m *MailSMTP) Send(ctx context.Context, p *EmailSend) error {
	if p == nil {
		return errors.New("email send parameters are nil")
	}
	if p.EmailFrom == "" || p.EmailTo == "" || p.EmailSubj == "" {
		return errors.New("required email fields are missing")
	}

	// Compose the email content
	headers := make(map[string]string)
	headers["From"] = p.EmailFrom
	headers["To"] = p.EmailTo
	headers["Subject"] = p.EmailSubj
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + p.EmailBody)

	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", m.Host, m.Port),
		m.Email,
		p.EmailFrom,
		[]string{p.EmailTo},
		[]byte(message.String()),
	)

	if err != nil {
		return err
	}

	return nil
}
