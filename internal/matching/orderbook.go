package matching

import (
	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"github.com/hokhoa05/core-engine/internal/models"
)

type InMemOrderBook struct {
	bids *redblacktree.Tree
	asks *redblacktree.Tree
}

func newInMemOrderBook() *InMemOrderBook {
	return &InMemOrderBook{
		bids: redblacktree.NewWith(func(a, b interface{}) int {
			return utils.UInt64Comparator(b, a)
		}),
		asks: redblacktree.NewWith(utils.UInt64Comparator),
	}
}

type IOrderBook interface {
	Add(order models.Order) error
	Cancel(order models.Order) error
	GetBestBid() (price uint64, qty uint64, err error)
	GetBestAsk() (price uint64, qty uint64, err error)
}
