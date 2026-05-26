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
		trades, _ := ob.Process(takerOrder)

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
		trades, _ := ob.Process(takerOrder)

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

// TestInMemOrderBook_ComplexPartialFills kiểm thử TSK-09
func TestInMemOrderBook_ComplexPartialFills(t *testing.T) {
	t.Run("Maker is partially filled by multiple takers until depleted", func(t *testing.T) {
		ob := NewInMemOrderBook()

		// GIVEN: 1 Maker bán 100 đơn vị giá 5000
		maker := models.Order{ID: 1, Side: models.Sell, Price: 5000, Qty: 100}
		_ = ob.Add(maker)

		// WHEN: Taker 1 mua 30 đơn vị
		taker1 := models.Order{ID: 2, Side: models.Buy, Price: 5000, Qty: 30}
		trades1, _ := ob.Process(taker1) // Bỏ qua error check cho gọn trong test này

		// THEN: Khớp 30. Maker còn 70.
		assert.Len(t, trades1, 1)
		assert.Equal(t, uint64(30), trades1[0].Qty)
		pl, _ := ob.asks.Get(uint64(5000))
		assert.Equal(t, uint64(70), pl.(*PriceLevel).Volume, "Level volume should drop to 70")

		// WHEN: Taker 2 mua tiếp 50 đơn vị
		taker2 := models.Order{ID: 3, Side: models.Buy, Price: 5000, Qty: 50}
		trades2, _ := ob.Process(taker2)

		// THEN: Khớp 50. Maker còn 20.
		assert.Len(t, trades2, 1)
		assert.Equal(t, uint64(50), trades2[0].Qty)
		assert.Equal(t, uint64(20), pl.(*PriceLevel).Volume, "Level volume should drop to 20")

		// WHEN: Taker 3 mua 40 đơn vị (Nhiều hơn số lượng còn lại)
		taker3 := models.Order{ID: 4, Side: models.Buy, Price: 5000, Qty: 40}
		trades3, _ := ob.Process(taker3)

		// THEN: Khớp nốt 20 của Maker. Taker 3 còn dư 20, tự động lên sổ làm Maker mới bên Bids.
		assert.Len(t, trades3, 1)
		assert.Equal(t, uint64(20), trades3[0].Qty, "Should only fill remaining 20")

		// Sổ Asks (Bán) phải bị xóa sạch do thanh khoản giá 5000 đã cạn
		assert.Equal(t, 0, ob.asks.Size(), "Asks tree should be completely purged")

		// Sổ Bids (Mua) phải chứa 20 đơn vị dư của Taker 3
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
