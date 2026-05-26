package matching

import (
	"container/list"
	"errors"

	"github.com/emirpasic/gods/trees/redblacktree"
	"github.com/emirpasic/gods/utils"
	"github.com/hokhoa05/core-engine/internal/models"
	"github.com/hokhoa05/core-engine/internal/pool"
)

type orderRef struct {
	element    *list.Element
	priceLevel *PriceLevel
}

type InMemOrderBook struct {
	bids           *redblacktree.Tree
	asks           *redblacktree.Tree
	tradePool      *pool.TradePool
	ordersRegistry map[uint64]orderRef
}

func NewInMemOrderBook() *InMemOrderBook {
	return &InMemOrderBook{
		bids: redblacktree.NewWith(func(a, b interface{}) int {
			return utils.UInt64Comparator(b, a)
		}),
		asks:           redblacktree.NewWith(utils.UInt64Comparator),
		ordersRegistry: make(map[uint64]orderRef),
		tradePool:      pool.NewTradePool(100000),
	}
}

func (ob *InMemOrderBook) Add(order models.Order) error {
	if err := order.Validate(false); err != nil {
		return err
	}
	if _, exists := ob.ordersRegistry[order.ID]; exists {
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

	ob.ordersRegistry[order.ID] = orderRef{
		element:    element,
		priceLevel: pl,
	}
	return nil
}

func (ob *InMemOrderBook) Cancel(orderID uint64) error {
	ref, exists := ob.ordersRegistry[orderID]
	if !exists {
		return errors.New("order not found")
	}

	pl := ref.priceLevel
	order := ref.element.Value.(models.Order)

	pl.Orders.Remove(ref.element)
	pl.Volume -= order.Qty

	delete(ob.ordersRegistry, orderID)

	if pl.Orders.Len() == 0 {
		if order.Side == models.Buy {
			ob.bids.Remove(order.Price)
		} else {
			ob.asks.Remove(order.Price)
		}
	}
	return nil
}

func (ob *InMemOrderBook) Process(taker models.Order, tradeBuffer *[]*models.Trade) error {
	if err := taker.Validate(false); err != nil {
		return err
	}
	if taker.Side == models.Buy {
		for taker.Qty > 0 && ob.asks.Size() > 0 {
			minNode := ob.asks.Left()
			bestAskPrice := minNode.Key.(uint64)

			if taker.Price < bestAskPrice {
				break
			}

			priceLevel := minNode.Value.(*PriceLevel)

			ob.matchWithPriceLevel(priceLevel, &taker, tradeBuffer)

			if priceLevel.Orders.Len() == 0 {
				ob.asks.Remove(bestAskPrice)
			}
		}
	} else {
		for taker.Qty > 0 && ob.bids.Size() > 0 {
			maxNode := ob.bids.Left()
			bestBidPrice := maxNode.Key.(uint64)

			if taker.Price > bestBidPrice {
				break
			}

			priceLevel := maxNode.Value.(*PriceLevel)

			ob.matchWithPriceLevel(priceLevel, &taker, tradeBuffer)

			if priceLevel.Orders.Len() == 0 {
				ob.bids.Remove(bestBidPrice)
			}
		}
	}
	if taker.Qty > 0 {
		_ = ob.Add(taker)
	}
	return nil
}

func (ob *InMemOrderBook) matchWithPriceLevel(pl *PriceLevel, taker *models.Order, tradeBuffer *[]*models.Trade) {
	currElem := pl.Orders.Front()

	for currElem != nil && taker.Qty > 0 {
		nextElem := currElem.Next()
		maker := currElem.Value.(models.Order)

		matchQty := min(taker.Qty, maker.Qty)
		tradeObj := ob.tradePool.Borrow()
		if tradeObj != nil {
			tradeObj.MakerOrderID = maker.ID
			tradeObj.TakerOrderID = taker.ID
			tradeObj.Price = pl.Price
			tradeObj.Qty = matchQty

			*tradeBuffer = append(*tradeBuffer, tradeObj)
		}

		taker.Qty -= matchQty
		maker.Qty -= matchQty
		pl.Volume -= matchQty

		if maker.Qty == 0 {
			pl.Orders.Remove(currElem)
			delete(ob.ordersRegistry, maker.ID)
		} else {
			currElem.Value = maker
		}
		currElem = nextElem
	}
}
func (ob *InMemOrderBook) ProcessMarketOrder(taker models.Order, tradeBuffer *[]*models.Trade) error {
	if err := taker.Validate(true); err != nil {
		return err
	}
	if taker.Side == models.Buy {
		for taker.Qty > 0 && ob.asks.Size() > 0 {
			minNode := ob.asks.Left()
			priceLevel := minNode.Value.(*PriceLevel)

			ob.matchWithPriceLevel(priceLevel, &taker, tradeBuffer)

			if priceLevel.Orders.Len() == 0 {
				ob.asks.Remove(minNode.Key)
			}
		}
	} else {
		for taker.Qty > 0 && ob.bids.Size() > 0 {
			maxNode := ob.bids.Left()
			priceLevel := maxNode.Value.(*PriceLevel)

			ob.matchWithPriceLevel(priceLevel, &taker, tradeBuffer)

			if priceLevel.Orders.Len() == 0 {
				ob.bids.Remove(maxNode.Key)
			}
		}
	}
	return nil
}
