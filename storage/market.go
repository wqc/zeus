package storage

import (
	"time"
)

const (
	BidType = 1
	AskType = 2
)

type Market struct {
	ID         int64     `orm:"column(id);pk"`
	Symbol     string    `orm:"column(symbol)"`
	Timestamp  int64     `orm:"column(ts)"`
	Type       int       `orm:"column(ptype)"`
	UpdateTime time.Time `orm:"column(update_time)"`
	Price      float64   `orm:"column(price)"`
	Quantity   float64   `orm:"column(quantity)"`
}
