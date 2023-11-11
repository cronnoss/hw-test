package storage

import (
	"time"
)

type Event struct {
	ID          int64
	UserID      int64
	Title       string
	Description string
	OnTime      time.Time
	OffTime     time.Time
	NotifyTime  time.Time
}
