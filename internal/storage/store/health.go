package store

import (
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
)

func NewHealthDB() *types.HealthDB {
	return &types.HealthDB{
		Default: types.ProcessorHealth{
			Failing:         false,
			MinResponseTime: 0,
			LastChecked:     time.Now().Truncate(time.Second),
		},
		Fallback: types.ProcessorHealth{
			Failing:         false,
			MinResponseTime: 0,
			LastChecked:     time.Now().Truncate(time.Second),
		},
	}
}
