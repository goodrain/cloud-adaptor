package models

import (
	"time"
)

type Model struct {
	ID        uint      `json:"-"`
	CreatedAt time.Time `json:"create_time"`
	UpdatedAt time.Time `json:"update_time"`
}
