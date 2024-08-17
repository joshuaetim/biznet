package main

import "time"

type createRecordRequest struct {
	Data      string
	OrderDate time.Time
	Forward   bool
	Location  string
}
