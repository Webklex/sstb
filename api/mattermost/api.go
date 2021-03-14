package mattermost

import (
	"../../utils/log"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"sync"
)

func NewMattermostApi(c *Config) *Config {
	cookieJar, _ := cookiejar.New(nil)

	return &Config{
		Endpoint: c.Endpoint,
		Token:    c.Token,
		Channel:  c.Channel,
		Username: c.Username,
		Header:   make(map[string]string),
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Jar: cookieJar,
		},
		mx: sync.Mutex{},
	}
}

func (c *Config) SendText(text string) {
	m := &Message{
		Channel: c.Channel,
		Text:    text,
	}
	c.Post(m)
}

func (c *Config) Post(v interface{}) []byte {
	c.mx.Lock()
	data, err := json.Marshal(v)
	req, err := http.NewRequest("POST", c.Endpoint+"/api/v4/posts", bytes.NewBuffer(data))
	if err != nil {
		log.Error(err)
		c.mx.Unlock()
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	for key, val := range c.Header {
		req.Header.Add(key, val)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		log.Error(err)
		c.mx.Unlock()
		return nil
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		_ = resp.Body.Close()
		log.Error(err)
		c.mx.Unlock()
		return nil
	}
	_ = resp.Body.Close()
	c.mx.Unlock()

	return bodyBytes
}
