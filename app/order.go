package app

import (
	"../utils/values"
	"time"
)

type Order struct {
	Id     int64         `json:"id"`
	Volume *values.Float `json:"volume"`
	Price  *values.Float `json:"price"`
	Total  *values.Float `json:"total"`
	Fee    *values.Float `json:"fee"`
	Side   string        `json:"side"`   // "sell" or "buy"
	Status string        `json:"status"` // "new", "filled", "canceled" or "other"
	Date   time.Time     `json:"date"`
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