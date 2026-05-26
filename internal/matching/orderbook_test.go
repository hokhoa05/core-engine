package matching

import (
	"testing"

	"github.com/hokhoa05/core-engine/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestInMemOrderBook_Add_And_Cancel(t *testing.T) {
	ob := NewInMemOrderBook()

	t.Run("Given a new Limit Buy Order, When added, Then it exists in Tree and Registry", func(t *testing.T) {
		order := models.Order{ID: 1, Side: models.Buy, Price: 10000, Qty: 50}
		err := ob.Add(order)

		assert.NoError(t, err)
		assert.Equal(t, 1, ob.bids.Size(), "Bids tree should have 1 price level")
		assert.Contains(t, ob.ordersRegistry, uint64(1), "Order should be in registry")
	})

	t.Run("Given an existing Order, When cancelled, Then memory is completely freed O(1)", func(t *testing.T) {
		err := ob.Cancel(1)

		assert.NoError(t, err)
		assert.Equal(t, 0, ob.bids.Size(), "Bids tree should be empty after last order is cancelled")
		assert.NotContains(t, ob.ordersRegistry, uint64(1), "Registry should be cleared")
	})

	t.Run("Given duplicate Order ID, When added, Then return error", func(t *testing.T) {
		order1 := models.Order{ID: 2, Side: models.Sell, Price: 10500, Qty: 10}
		_ = ob.Add(order1)

		order2 := models.Order{ID: 2, Side: models.Sell, Price: 10600, Qty: 20}
		err := ob.Add(order2) // Cố tình thêm trùng ID

		assert.Error(t, err)
		assert.Equal(t, "order ID already exists", err.Error())
	})
}

func TestInMemOrderBook_SimpleMatch(t *testing.T) {
	t.Run("Exact Match: 1 Taker completely fills 1 Maker", func(t *testing.T) {
		ob := NewInMemOrderBook()

		makerOrder := models.Order{ID: 101, Side: models.Sell, Price: 15000, Qty: 100}
		_ = ob.Add(makerOrder)

		takerOrder := models.Order{ID: 201, Side: models.Buy, Price: 15000, Qty: 100}
		trades, _ := ob.Process(takerOrder)

		assert.Len(t, trades, 1, "Should generate exactly 1 trade")
		assert.Equal(t, uint64(15000), trades[0].Price)
		assert.Equal(t, uint64(100), trades[0].Qty)
		assert.Equal(t, uint64(101), trades[0].MakerOrderID)
		assert.Equal(t, uint64(201), trades[0].TakerOrderID)

		assert.Equal(t, 0, ob.asks.Size())
		assert.Empty(t, ob.ordersRegistry)
	})

	t.Run("Partial Match: Taker buys less than Maker offers", func(t *testing.T) {
		ob := NewInMemOrderBook()

		makerOrder := models.Order{ID: 102, Side: models.Sell, Price: 15000, Qty: 100}
		_ = ob.Add(makerOrder)

		takerOrder := models.Order{ID: 202, Side: models.Buy, Price: 15000, Qty: 40}
		trades, _ := ob.Process(takerOrder)

		assert.Len(t, trades, 1)
		assert.Equal(t, uint64(40), trades[0].Qty)

		pl, _ := ob.asks.Get(uint64(15000))
		assert.Equal(t, uint64(60), pl.(*PriceLevel).Volume, "Maker should have 60 qty remaining")
	})
}

func TestInMemOrderBook_MarketOrder(t *testing.T) {
	t.Run("Market Buy Order sweeps multiple ask levels and discards remaining qty", func(t *testing.T) {
		ob := NewInMemOrderBook()

		_ = ob.Add(models.Order{ID: 101, Side: models.Sell, Price: 100, Qty: 10}) // Rẻ hơn
		_ = ob.Add(models.Order{ID: 102, Side: models.Sell, Price: 105, Qty: 15}) // Đắt hơn

		takerOrder := models.Order{ID: 201, Side: models.Buy, Price: 0, Qty: 30} // Giá 0 vì là Market Order
		trades := ob.ProcessMarketOrder(takerOrder)

		assert.Len(t, trades, 2, "Should sweep 2 price levels")

		assert.Equal(t, uint64(10), trades[0].Qty)
		assert.Equal(t, uint64(100), trades[0].Price)

		assert.Equal(t, uint64(15), trades[1].Qty)
		assert.Equal(t, uint64(105), trades[1].Price)

		assert.Equal(t, 0, ob.asks.Size())

		assert.Equal(t, 0, ob.bids.Size(), "Market order remainder should NOT be added to book")
	})
}

func TestInMemOrderBook_ComplexPartialFills(t *testing.T) {
	t.Run("Maker is partially filled by multiple takers until depleted", func(t *testing.T) {
		ob := NewInMemOrderBook()

		maker := models.Order{ID: 1, Side: models.Sell, Price: 5000, Qty: 100}
		_ = ob.Add(maker)

		taker1 := models.Order{ID: 2, Side: models.Buy, Price: 5000, Qty: 30}
		trades1, _ := ob.Process(taker1) // Bỏ qua error check cho gọn trong test này

		assert.Len(t, trades1, 1)
		assert.Equal(t, uint64(30), trades1[0].Qty)
		pl, _ := ob.asks.Get(uint64(5000))
		assert.Equal(t, uint64(70), pl.(*PriceLevel).Volume, "Level volume should drop to 70")

		taker2 := models.Order{ID: 3, Side: models.Buy, Price: 5000, Qty: 50}
		trades2, _ := ob.Process(taker2)

		assert.Len(t, trades2, 1)
		assert.Equal(t, uint64(50), trades2[0].Qty)
		assert.Equal(t, uint64(20), pl.(*PriceLevel).Volume, "Level volume should drop to 20")

		taker3 := models.Order{ID: 4, Side: models.Buy, Price: 5000, Qty: 40}
		trades3, _ := ob.Process(taker3)

		assert.Len(t, trades3, 1)
		assert.Equal(t, uint64(20), trades3[0].Qty, "Should only fill remaining 20")

		assert.Equal(t, 0, ob.asks.Size(), "Asks tree should be completely purged")

		assert.Equal(t, 1, ob.bids.Size())
		bidPl, _ := ob.bids.Get(uint64(5000))
		assert.Equal(t, uint64(20), bidPl.(*PriceLevel).Volume)
	})

	t.Run("Zero Quantity Order is rejected", func(t *testing.T) {
		ob := NewInMemOrderBook()
		order := models.Order{ID: 99, Side: models.Buy, Price: 100, Qty: 0}

		err := ob.Add(order)
		assert.Error(t, err)
		assert.Equal(t, "order quantity must be greater than zero", err.Error())
	})
}
