package notifier

import (
	"../../api/mattermost"
	"../../api/slack"
	"../../utils/log"
	"sync"
)

type Notifier struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
	Channel  string `json:"channel"`
	Username string `json:"username"`
	Driver   string `json:"driver"`
	Name     string `json:"name"`
	Email    *Email `json:"email"`

	mx sync.Mutex         `json:"-"`
	mm *mattermost.Config `json:"-"`
}

func (n *Notifier) Send(msg string) {
	if n.Driver == "mattermost" {
		n.mm.SendText(msg)
	} else if n.Driver == "email" {
		n.Email.Send(msg)
	} else if n.Driver == "slack" {
		if err := slack.SendNotification(n.Endpoint, msg); err != nil {
			log.Error(err)
		}
	}
}

func (n *Notifier) Init() {
	if n.Driver == "mattermost" {
		n.mm = mattermost.NewMattermostApi(&mattermost.Config{
			Endpoint: n.Endpoint,
			Token:    n.Token,
			Channel:  n.Channel,
			Username: n.Username,
		})
	}
}
