package matching

import "github.com/hokhoa05/core-engine/internal/models"

type IMatchingEngine interface {
	Process(order models.Order, tradeBuffer *[]*models.Trade) error
	ProcessMarketOrder(order models.Order, tradeBuffer *[]*models.Trade) error
	Add(order models.Order) error
	Cancel(orderID uint64) error
}
