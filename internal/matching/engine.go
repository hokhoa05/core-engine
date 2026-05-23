package matching

import "github.com/hokhoa05/core-engine/internal/models"

type IMatchingEngine interface {
	Process(order models.Order) error
}
