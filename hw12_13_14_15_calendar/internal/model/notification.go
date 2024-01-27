package model

import "time"

type NotificationMsg struct {
	ID     int64
	Title  string
	Date   time.Time
	UserID int64
}
