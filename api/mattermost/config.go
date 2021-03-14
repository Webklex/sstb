package mattermost

import (
	"net/http"
	"sync"
)

type Config struct {
	Endpoint string `json:"endpoint"`
	Token    string `json:"token"`
	Channel  string `json:"channel"`
	Username string `json:"username"`

	Header map[string]string
	Client *http.Client

	mx sync.Mutex `json:"-"`
}

type Message struct {
	Channel  string `json:"channel_id"`
	Text     string `json:"message"`
}
