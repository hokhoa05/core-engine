package matching

import (
	"fmt"

	"github.com/hokhoa05/core-engine/internal/models"
)

type CommandType int

const (
	CmdPlaceOlder CommandType = iota
	CmdCancelOrder
)

type Command struct {
	Type    CommandType
	Order   models.Order
	OrderID uint64
}

type EngineRunner struct {
	orderBook   *InMemOrderBook
	commandChan chan Command
	tradeBuffer []*models.Trade
}

func NewEngineRunner(bufferSize int) *EngineRunner {
	return &EngineRunner{
		orderBook:   NewInMemOrderBook(),
		commandChan: make(chan Command, bufferSize),
		tradeBuffer: make([]*models.Trade, 0, 100),
	}
}

func (r *EngineRunner) Start() {
	fmt.Println("Matching Engine Event Loop Started...")

	for cmd := range r.commandChan {
		switch cmd.Type {
		case CmdPlaceOlder:
			r.tradeBuffer = r.tradeBuffer[:0]
			if cmd.Order.Price == 0 {
				_ = r.orderBook.ProcessMarketOrder(cmd.Order, &r.tradeBuffer)
			} else {
				_ = r.orderBook.Process(cmd.Order, &r.tradeBuffer)
			}

			for _, trade := range r.tradeBuffer {
				fmt.Printf("[TRADE MATCHED] Maker: %d | Taker: %d | Price: %d | Qty: %d\n",
					trade.MakerOrderID, trade.TakerOrderID, trade.Price, trade.Qty)

				r.orderBook.tradePool.Return()
			}
		case CmdCancelOrder:
			err := r.orderBook.Cancel(cmd.OrderID)
			if err != nil {
				fmt.Printf("[CANCEL ERROR] OrderID %d: %v\n", cmd.OrderID, err)
			} else {
				fmt.Printf("[ORDER CANCELLED] OrderID %d\n", cmd.OrderID)
			}
		}
	}
}

func (r *EngineRunner) PushCommand(cmd Command) {
	r.commandChan <- cmd
}
