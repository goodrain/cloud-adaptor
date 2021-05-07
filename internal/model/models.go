package model

import (
	"time"
)

// Model -
type Model struct {
	ID        uint      `json:"-"`
	CreatedAt time.Time `json:"create_time"`
	UpdatedAt time.Time `json:"update_time"`
}
