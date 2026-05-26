package models

import "errors"

type Side int8

const (
	Buy  Side = 1
	Sell Side = -1
)

type Order struct {
	ID    uint64
	Price uint64
	Qty   uint64
	Side  Side
}

func (o *Order) Validate(isMarketOrder bool) error {
	if o.Qty <= 0 {
		return errors.New("order quantity must be greater than zero")
	}
	if o.Side != Sell && o.Side != Buy {
		return errors.New("Invalid order side")
	}
	if !isMarketOrder && o.Price <= 0 {
		return errors.New("limit order price must be greater than zero")
	}
	return nil
}

type Trade struct {
	MakerOrderID uint64
	TakerOrderID uint64
	Price        uint64
	Qty          uint64
}
