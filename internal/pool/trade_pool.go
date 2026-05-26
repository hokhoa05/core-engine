package pool

import "github.com/hokhoa05/core-engine/internal/models"

type TradePool struct {
	buffer   []models.Trade
	capacity int
	head     int
	tail     int
	count    int
}

func NewTradePool(size int) *TradePool {
	return &TradePool{
		buffer:   make([]models.Trade, size),
		capacity: size,
		head:     0,
		tail:     0,
		count:    0,
	}
}

func (p *TradePool) Borrow() *models.Trade {
	if p.count == p.capacity {
		return nil
	}

	trade := &p.buffer[p.head]
	p.head = (p.head + 1) % p.capacity
	p.count++

	return trade
}

func (p *TradePool) Return() {
	if p.count == 0 {
		return
	}

	p.tail = (p.tail + 1) % p.capacity
	p.count--
}
