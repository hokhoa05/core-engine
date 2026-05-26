package matching

import (
	"testing"

	"github.com/hokhoa05/core-engine/internal/models"
)

func BenchmarkInMemoryBook_Add(b *testing.B) {
	ob := NewInMemOrderBook()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		order := models.Order{
			ID:    uint64(i),
			Side:  models.Buy,
			Price: 5000,
			Qty:   10,
		}

		_ = ob.Add(order)
	}
}

func BenchmarkInMemoryBook_Process(b *testing.B) {
	ob := NewInMemOrderBook()

	ob.Add(models.Order{
		ID:    1,
		Side:  models.Sell,
		Price: 5000,
		Qty:   uint64(b.N) * 10,
	})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		taker := models.Order{
			ID:    uint64(i + 2),
			Side:  models.Buy,
			Price: 5000,
			Qty:   10,
		}
		var trades []*models.Trade
		_ = ob.Process(taker, &trades)
	}
}
