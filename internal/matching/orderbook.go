package matching

import "github.com/hokhoa05/core-engine/internal/models"

type IOrderBook interface {
	Add(order models.Order) error
	Cancel(order models.Order) error
	GetBestBid() (price uint64, qty uint64, err error)
	GetBestAsk() (price uint64, qty uint64, err error)
}
