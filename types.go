package main

import "time"

type createRecordRequest struct {
	Data      string    `json:"data"`
	OrderDate time.Time `json:"order_date"`
	Forward   bool      `json:"forward"`
	Location  string    `json:"location"`
}
