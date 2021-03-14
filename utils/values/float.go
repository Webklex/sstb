package values

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type Float struct {
	big.Float
}

const PRECISION = 8

var ZeroFloat = NewEmptyFloat()
var HundredFloat = NewFloatFromFloat64(100)

func NewFloatFromFloat64(x float64) *Float {
	fl := big.NewFloat(x)
	return NewFloat(fl)
}

func NewEmptyFloat() *Float {
	return NewFloatFromFloat64(0)
}

func NewFloat(x *big.Float) *Float {
	return &Float{
		Float: *x,
	}
}

func NewFloatFromString(x string) *Float {
	s := strings.Trim(x, "\"")
	f := &Float{}
	if s == "null" {
		f.Float = big.Float{}
		return f
	}
	if fl, ok := new(big.Float).SetString(s); ok {
		f.Float = *fl
	}

	return f
}

func (f *Float) ToFloat() float64 {
	fl, _ := f.Float64()
	return fl
}

func (f *Float) Gt(other *Float) bool {
	return f.Float.Cmp(&other.Float) > 0
}

func (f *Float) Lt(other *Float) bool {
	return f.Float.Cmp(&other.Float) < 0
}

func (f *Float) Eq(other *Float) bool {
	return f.Float.Cmp(&other.Float) == 0
}

func (f *Float) UnmarshalJSON(b []byte) (err error) {
	f.Float = NewFloatFromString(string(b)).Float
	return
}

func (f *Float) MarshalJSON() ([]byte, error) {
	if f.IsInf() {
		return []byte("null"), nil
	}
	return []byte("\"" + f.ToString() + "\""), nil
}

func (f *Float) ToString() string {
	return f.ToPrecision(PRECISION)
}

func (f *Float) ToPrecision(pre int) string {
	return fmt.Sprintf("%."+strconv.Itoa(pre)+"f", f.ToFloat())
}

func (f *Float) Difference(other *Float) *Float {
	if other.String() != f.Float.String() {
		return other.Sub(f).Div(f).Mul(NewFloatFromFloat64(100))
	}
	return NewEmptyFloat()
}

func (f *Float) Add(other *Float) *Float {
	delta := new(big.Float).Add(&f.Float, &other.Float)
	return NewFloat(delta)
}

func (f *Float) Sub(other *Float) *Float {
	delta := new(big.Float).Sub(&f.Float, &other.Float)
	return NewFloat(delta)
}

func (f *Float) Mul(other *Float) *Float {
	delta := new(big.Float).Mul(&f.Float, &other.Float)
	return NewFloat(delta)
}

func (f *Float) Quo(other *Float) *Float {
	delta := new(big.Float).Quo(&f.Float, &other.Float)
	return NewFloat(delta)
}

func (f *Float) Div(other *Float) *Float {
	z := NewEmptyFloat()
	if f.Eq(z) || other.Eq(z) {
		return z
	}
	return f.Quo(other)
}
