package matching

import (
	"testing"

	"github.com/hokhoa05/core-engine/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestInMemOrderBook_Add_And_Cancel kiểm thử TSK-04 và TSK-05
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

// TestInMemOrderBook_SimpleMatch kiểm thử TSK-06 (Core Matching Logic)
func TestInMemOrderBook_SimpleMatch(t *testing.T) {
	t.Run("Exact Match: 1 Taker completely fills 1 Maker", func(t *testing.T) {
		ob := NewInMemOrderBook()

		// GIVEN: Sổ lệnh có 1 người đang rải lệnh Bán (Maker)
		makerOrder := models.Order{ID: 101, Side: models.Sell, Price: 15000, Qty: 100}
		_ = ob.Add(makerOrder)

		// WHEN: Một người Mua (Taker) vào mua đúng giá và số lượng đó
		takerOrder := models.Order{ID: 201, Side: models.Buy, Price: 15000, Qty: 100}
		trades := ob.Process(takerOrder)

		// THEN: Khớp 100%. Sổ lệnh trống trơn.
		assert.Len(t, trades, 1, "Should generate exactly 1 trade")
		assert.Equal(t, uint64(15000), trades[0].Price)
		assert.Equal(t, uint64(100), trades[0].Qty)
		assert.Equal(t, uint64(101), trades[0].MakerOrderID)
		assert.Equal(t, uint64(201), trades[0].TakerOrderID)

		// Kiểm tra bộ nhớ đã dọn dẹp sạch sẽ
		assert.Equal(t, 0, ob.asks.Size())
		assert.Empty(t, ob.ordersRegistry)
	})

	t.Run("Partial Match: Taker buys less than Maker offers", func(t *testing.T) {
		ob := NewInMemOrderBook()

		// GIVEN: Maker bán 100
		makerOrder := models.Order{ID: 102, Side: models.Sell, Price: 15000, Qty: 100}
		_ = ob.Add(makerOrder)

		// WHEN: Taker chỉ mua 40
		takerOrder := models.Order{ID: 202, Side: models.Buy, Price: 15000, Qty: 40}
		trades := ob.Process(takerOrder)

		// THEN: Lệnh Maker vẫn còn lại 60 trên sổ
		assert.Len(t, trades, 1)
		assert.Equal(t, uint64(40), trades[0].Qty)

		pl, _ := ob.asks.Get(uint64(15000))
		assert.Equal(t, uint64(60), pl.(*PriceLevel).Volume, "Maker should have 60 qty remaining")
	})
}

// TestInMemOrderBook_MarketOrder kiểm thử TSK-08
func TestInMemOrderBook_MarketOrder(t *testing.T) {
	t.Run("Market Buy Order sweeps multiple ask levels and discards remaining qty", func(t *testing.T) {
		ob := NewInMemOrderBook()

		// GIVEN: Sổ lệnh có 2 mức giá bán khác nhau
		_ = ob.Add(models.Order{ID: 101, Side: models.Sell, Price: 100, Qty: 10}) // Rẻ hơn
		_ = ob.Add(models.Order{ID: 102, Side: models.Sell, Price: 105, Qty: 15}) // Đắt hơn

		// WHEN: Lệnh Market Buy muốn mua tới 30 đơn vị (Nhiều hơn tổng thanh khoản là 25)
		takerOrder := models.Order{ID: 201, Side: models.Buy, Price: 0, Qty: 30} // Giá 0 vì là Market Order
		trades := ob.ProcessMarketOrder(takerOrder)

		// THEN: Khớp sinh ra 2 trades ở 2 mức giá khác nhau
		assert.Len(t, trades, 2, "Should sweep 2 price levels")

		// Trade 1: Khớp hết 10 đơn vị ở giá 100
		assert.Equal(t, uint64(10), trades[0].Qty)
		assert.Equal(t, uint64(100), trades[0].Price)

		// Trade 2: Khớp nốt 15 đơn vị ở giá 105 (Hiện tượng trượt giá)
		assert.Equal(t, uint64(15), trades[1].Qty)
		assert.Equal(t, uint64(105), trades[1].Price)

		// Kiểm tra thanh khoản đã bị quét sạch
		assert.Equal(t, 0, ob.asks.Size())

		// Lệnh Market Order mua 30, chỉ khớp 25, thiếu 5. Nhưng 5 đơn vị này không được lên sổ Mua.
		assert.Equal(t, 0, ob.bids.Size(), "Market order remainder should NOT be added to book")
	})
}
