package worker

import (
	"log"

	"github.com/hokhoa05/core-engine/internal/client"
	"github.com/hokhoa05/core-engine/internal/models"
)

type SettlementWorker struct {
	client       *client.SettlementClient
	tradeChannel <-chan models.Trade
}

func NewSettlementWorker(c *client.SettlementClient, ch <-chan models.Trade) *SettlementWorker {
	return &SettlementWorker{
		client:       c,
		tradeChannel: ch,
	}
}

func (w *SettlementWorker) Start() {
	log.Println("Settlement Worker is initialized, waiting...")

	for trade := range w.tradeChannel {
		err := w.client.ReportTrade(
			trade.MakerOrderID,
			trade.TakerOrderID,
			trade.MakerUserID,
			trade.TakerUserID,
			trade.Price,
			trade.Qty,
		)

		if err != nil {
			log.Printf("WARNING: Java is down. Need to retry Trade Maker %d - Taker %d", trade.MakerOrderID, trade.TakerOrderID)
		}
	}
}
