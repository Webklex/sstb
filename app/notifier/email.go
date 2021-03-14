package notifier

import (
	"../../utils/log"
	"github.com/scorredoira/email"
	"net/mail"
	"net/smtp"
)

type Email struct {
	To      []string `json:"to"`
	Cc      []string `json:"cc"`
	Bcc     []string `json:"bcc"`
	Subject string   `json:"subject"`
	Name    string   `json:"name"`
	Sender  string   `json:"sender"`

	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (e *Email) Send(msg string) {
	if (len(e.To) + len(e.Cc) + len(e.Bcc)) > 0 {
		m := email.NewMessage(e.Subject, msg)
		m.From = mail.Address{Name: e.Name, Address: e.Sender}
		m.To = e.To
		m.Cc = e.Cc
		m.Bcc = e.Bcc

		auth := smtp.PlainAuth(e.Sender, e.Username, e.Password, e.Host)
		if err := email.Send(e.Host+":"+e.Port, auth, m); err != nil {
			log.Error(err)
		}
	}
}
