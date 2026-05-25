package matching

import (
	"container/list"

	"github.com/hokhoa05/core-engine/internal/models"
)

type PriceLevel struct {
	Price  uint64
	Orders *list.List
	Volume uint64
}

func NewPriceLevel(price uint64) *PriceLevel {
	return &PriceLevel{
		Price:  price,
		Orders: list.New(),
		Volume: 0,
	}
}

func (pl *PriceLevel) Append(order models.Order) *list.Element {
	pl.Volume += order.Qty
	return pl.Orders.PushBack(order)
}
