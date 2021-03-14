package values

import (
	"strings"
)

type String struct {
	Value string
}

func (c *String) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	c.Value = s
	return
}

func (c *String) MarshalJSON() ([]byte, error) {
	return []byte("\"" + c.Value + "\""), nil
}
