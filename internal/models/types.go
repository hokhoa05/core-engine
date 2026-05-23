package models

type Order struct {
	ID    string
	Price uint64
	Qty   uint64
	Side  int
}
