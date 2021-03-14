package log

import (
	"../config"
	"../filesystem"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/kelseyhightower/envconfig"
	"io"
	"io/ioutil"
	olog "log"
	"os"
	"path"
	"strconv"
	"time"
)

var Log = DefaultConfig()

func DefaultConfig() *Config {
	dir, _ := os.Getwd()

	c := &Config{
		Config:        config.DefaultConfig(),
		LogToStdout:   false,
		LogTimestamp:  true,
		Debug:         false,
		LogOutputFile: path.Join(dir, "logs", "log_"+strconv.Itoa(int(time.Now().Unix()))+".log"),
		Silent:        true,
		logger:        make(map[string]*olog.Logger),
	}
	c.File = path.Join(dir, "config", "log.json")
	c.Config.SetContext(c)

	return c
}

func NewConfigFromFile(configFile string) *Config {
	c := DefaultConfig()

	c.Load(configFile)
	c.File = configFile

	return c
}

// AddFlags adds configuration flags to the given FlagSet.
func (c *Config) AddFlags(fs *flag.FlagSet) {
	defer envconfig.Process("sstb", c)

	fs.BoolVar(&c.LogToStdout, "logtostdout", c.LogToStdout, "Log to stdout instead of stderr")
	fs.StringVar(&c.LogOutputFile, "output-file", c.LogOutputFile, "Log output file")
	fs.BoolVar(&c.LogTimestamp, "logtimestamp", c.LogTimestamp, "Prefix non-access logs with timestamp")

	fs.BoolVar(&c.Silent, "silent", c.Silent, "Disable logging and suppress any output")
	fs.BoolVar(&c.Debug, "debug", c.Debug, "Enable the debug mode")
	fs.StringVar(&c.File, "log-config", c.File, "Log config file")
}

func (c *Config) Init() {

	c.LogOutput = os.Stdout

	if c.LogToStdout {
		olog.SetOutput(c.LogOutput)
	} else if c.LogOutputFile != "" {
		_, _ = filesystem.MakeDir(c.LogOutputFile)
		_ = ioutil.WriteFile(c.LogOutputFile, []byte(""), 0644)
		c.LogOutput, _ = os.OpenFile(c.LogOutputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		olog.SetOutput(c.LogOutput)
	}
	if !c.LogTimestamp {
		olog.SetFlags(0)
	}

	lflag := 0
	if c.LogTimestamp {
		lflag = olog.LstdFlags
	}
	c.logger["error"] = olog.New(c.logWriter(), "[error] ", lflag)
	c.logger["info"] = olog.New(c.logWriter(), "[info] ", lflag)
	c.logger["update"] = olog.New(c.logWriter(), "[update] ", lflag)
	c.logger["access"] = olog.New(c.logWriter(), "[access] ", lflag)
	c.logger["other"] = olog.New(c.logWriter(), "[other] ", lflag)
	c.logger["debug"] = olog.New(c.logWriter(), "[debug] ", lflag)
	c.logger["success"] = olog.New(c.logWriter(), "[success] ", lflag)
	c.logger["warning"] = olog.New(c.logWriter(), "[warning] ", lflag)
}

func (c *Config) log(_type string, a ...interface{}) {
	if c.Silent {
		return
	}

	_datetime := time.Now().Format("2006/01/02 15:04:05")
	message := fmt.Sprintf("%s %s", _datetime, a[0])
	switch _type {
	case "error":
		color.Magenta(message)
	case "info":
		color.Cyan(message)
	case "success":
		color.Green(message)
	case "access":
		color.Black(message)
	case "warning":
		color.Yellow(message)
	default:
		color.Cyan(message)
		_type = "other"
	}

	if logger, ok := c.logger[_type]; ok && c.LogOutputFile != "" {
		logger.Println(a[0])
	}
}

func Success(v ...interface{}) {
	Log.log("success", v)
}

func Access(v ...interface{}) {
	Log.log("access", v)
}

func Println(v ...interface{}) {
	Log.log("other", v)
}

func Debug(v ...interface{}) {
	if Log.Debug {
		Log.log("debug", v)
	}
}

func Info(v ...interface{}) {
	Log.log("info", v)
}

func Warn(v ...interface{}) {
	Log.log("warning", v)
}

func Error(v ...interface{}) {
	Log.log("error", v)
}

func Fatal(v ...interface{}) {
	Error(v)
	os.Exit(1)
}

func (c *Config) logWriter() io.Writer {
	return c.LogOutput
}

func (c *Config) ErrorLogger() *olog.Logger {
	return c.logger["error"]
}

func (c *Config) AccessLogger() *olog.Logger {
	return c.logger["access"]
}
