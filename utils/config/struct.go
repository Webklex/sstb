package config

type Config struct {
	File    string `json:"-"`
	RootDir string `json:"-"`
	Silent  bool   `json:"-"`

	context interface{} `json:"-"`
}

type CtxKey struct{}
