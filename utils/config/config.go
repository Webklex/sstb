package config

import (
	"../filesystem"
	"encoding/json"
	"fmt"
	"github.com/oleiade/reflections"
	"io/ioutil"
	"log"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func NewConfig() *Config {
	return &Config{}
}

func DefaultConfig() *Config {
	dir, _ := os.Getwd()

	c := &Config{
		RootDir: dir,
		File:    path.Join(dir, "config", "config.json"),
	}
	return c
}

func NewConfigFromFile(configFile string) *Config {
	config := DefaultConfig()

	if configFile != "" {
		config.Load(configFile)
		config.File = configFile
	}

	return config
}

func (c *Config) SetContext(ctx interface{}) {
	c.context = ctx
}

func (c *Config) initFile(filename string) {
	if len(filename) == 0 {
		dir, _ := os.Getwd()
		filename = path.Join(dir, "config", "config.json")

		c.Load(filename)
	} else {
		_, _ = filesystem.MakeDir(filename)
	}
	c.File = filename
}

func (c *Config) Load(filename string) bool {
	c.initFile(filename)

	if _, err := os.Stat(filename); err == nil {
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			if !c.Silent {
				log.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		err = json.Unmarshal(content, c.context)
		if err != nil {
			if !c.Silent {
				log.Printf("[error] Config file failed to load: %s", err.Error())
			}
			return false
		}

		if !c.Silent {
			log.Printf("[info] Config file loaded successfully")
		}

	} else {
		_, _ = c.Save()
	}

	return true
}

func (c *Config) Save() (bool, error) {
	if len(c.File) == 0 {
		c.initFile("")
	}

	file, err := json.MarshalIndent(c.context, "", "\t")
	if err != nil {
		if !c.Silent {
			log.Printf("[error] Config file failed to save: %s", err.Error())
		}
		return false, err
	} else if string(file) == "" {
		return true, nil
	}

	err = ioutil.WriteFile(c.File, file, 0644)
	if err != nil {
		if !c.Silent {
			log.Printf("[error] Config file failed to save: %s", err.Error())
		}
		return false, err
	}

	if !c.Silent {
		log.Printf("[info] Config file saved under: %s", c.File)
	}

	return true, nil
}

func (c *Config) Update(v interface{}, values map[string]interface{}) interface{} {

	val := reflect.ValueOf(v)
	ival := reflect.Indirect(val)

	for i := 0; i < ival.Type().NumField(); i++ {
		t := ival.Type().Field(i)
		fieldName := t.Name
		columnName := t.Name

		if jsonTag := t.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
			if commaIdx := strings.Index(jsonTag, ","); commaIdx > 0 {
				columnName = jsonTag[:commaIdx]
			} else {
				columnName = jsonTag
			}

			if value, ok := values[columnName]; ok {
				switch t.Type.Name() {
				case "int":
					_val, _ := strconv.Atoi(value.(string))
					_ = reflections.SetField(v, fieldName, _val)
				case "float64":
					_ = reflections.SetField(v, fieldName, value.(float64))
				case "bool":
					if (value.(string) != "false" && value.(string) != "null" && value.(string) != "disabled") && reflect.TypeOf(value).String() == "string" {
						_ = reflections.SetField(v, fieldName, true)
					} else {
						_ = reflections.SetField(v, fieldName, value)
					}
				case "string":
					_ = reflections.SetField(v, fieldName, value.(string))
				case "Duration":
					_time, _ := time.ParseDuration(value.(string))
					_ = reflections.SetField(v, fieldName, _time)
				case "[]byte":
					_ = reflections.SetField(v, fieldName, value.([]byte))
				default:
					fmt.Println("unknown", t, t.Type.Name())
					_ = reflections.SetField(v, fieldName, value.(string))
				}
			}
		}
	}

	return v
}
