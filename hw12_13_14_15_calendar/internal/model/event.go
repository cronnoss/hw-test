package model

import (
	"time"
)

type Event struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"userid"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	OnTime      time.Time `json:"ontime"`
	OffTime     time.Time `json:"offtime"`
	NotifyTime  time.Time `json:"notifytime,omitempty"`
	Notified    bool      `json:"-"`
}
