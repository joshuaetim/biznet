package main

import (
	"time"

	"github.com/google/uuid"
)

type Record struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Data      string    `json:"data"`
	OrderDate time.Time `json:"order_date"`
	Forward   bool      `json:"forward"`
	Location  string    `json:"location"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Record) TableName() string {
	return "records"
}
