package app

import (
	"../utils/values"
	"time"
)

type Order struct {
	Id     int64
	Volume *values.Float
	Price  *values.Float
	Total  *values.Float
	Fee    *values.Float
	Side   string // "sell" or "buy"
	Status string // "new", "filled", "canceled" or "other"
	Date   time.Time
}

func NewDefaultOrder() *Order {
	return &Order{
		Id:     0,
		Volume: values.NewEmptyFloat(),
		Price:  values.NewEmptyFloat(),
		Total:  values.NewEmptyFloat(),
		Fee:    values.NewEmptyFloat(),
		Side:   "",
		Date:   time.Time{},
	}
}
