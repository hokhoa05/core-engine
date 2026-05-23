package models

type Side int8

const (
	Buy Side = 1
	Sell Side = -1
)

type Order struct {
	ID    string
	Price uint64
	Qty   uint64
	Side  Side
}

type Trade struct {
	MakeOrderID uint64
	TakerOrderID uint64
	Price uint64
	Qty uint64
}
