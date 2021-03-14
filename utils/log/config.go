package log

import (
	"../config"
	olog "log"
	"os"
)

type Config struct {
	*config.Config

	Silent bool `json:"silent"`
	Debug  bool `json:"debug"`

	LogToStdout   bool   `json:"stdout"`
	LogOutputFile string `json:"file"`
	LogTimestamp  bool   `json:"timestamp"`

	LogOutput *os.File                `json:"-"`
	logger    map[string]*olog.Logger `json:"-"`
}
