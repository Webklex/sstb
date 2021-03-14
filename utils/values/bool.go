package values

import (
	"strings"
)

type Bool struct {
	State bool
}

func (c *Bool) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "false" || s == "0" {
		c.State = false
		return
	}
	c.State = true
	return
}

func (c *Bool) MarshalJSON() ([]byte, error) {
	if c.State == false {
		return []byte("false"), nil
	}
	return []byte("true"), nil
}
