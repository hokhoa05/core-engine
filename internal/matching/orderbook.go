package matching

import (
	"container/list"
	"errors"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"github.com/hokhoa05/core-engine/internal/models"
)

type orderRef struct {
	element    *list.Element
	priceLevel *PriceLevel
}

type InMemOrderBook struct {
	bids *redblacktree.Tree
	asks *redblacktree.Tree

	orderRegistry map[string]orderRef
}

func newInMemOrderBook() *InMemOrderBook {
	return &InMemOrderBook{
		bids: redblacktree.NewWith(func(a, b interface{}) int {
			return utils.UInt64Comparator(b, a)
		}),
		asks:          redblacktree.NewWith(utils.UInt64Comparator),
		orderRegistry: make(map[string]orderRef),
	}
}

func (ob *InMemOrderBook) Add(order models.Order) error {
	if _, exists := ob.orderRegistry[order.ID]; exists {
		return errors.New("order ID already exists")
	}

	tree := ob.asks
	if order.Side == models.Buy {
		tree = ob.bids
	}

	var pl *PriceLevel

	if value, found := tree.Get(order.Price); found {
		pl = value.(*PriceLevel)
	} else {
		pl = NewPriceLevel(order.Price)
		tree.Put(order.Price, pl)
	}

	element := pl.Append(order)

	ob.orderRegistry[order.ID] = orderRef{
		element:    element,
		priceLevel: pl,
	}
	return nil
}

func (ob *InMemOrderBook) Cancel(orderID string) error {
	ref, exists := ob.orderRegistry[orderID]
	if !exists {
		return errors.New("order not found")
	}

	pl := ref.priceLevel
	order := ref.element.Value.(models.Order)

	pl.Orders.Remove(ref.element)
	pl.Volume -= order.Qty

	delete(ob.orderRegistry, orderID)

	if pl.Orders.Len() == 0 {
		if order.Side == models.Buy {
			ob.bids.Remove(order.Price)
		} else {
			ob.asks.Remove(order.Price)
		}
	}
	return nil
}
