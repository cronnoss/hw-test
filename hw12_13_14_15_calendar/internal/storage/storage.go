package storage

import (
	"context"
	"time"
)

type Storage interface {
	Connect(context.Context) error
	Close(context.Context) error
	InsertEvent(context.Context, *Event) error
	UpdateEvent(context.Context, *Event) error
	DeleteEvent(context.Context, int64) error
	GetEventByID(context.Context, int64) (Event, error)
	GetAllEvents(context.Context, int64) ([]Event, error)
	GetAllRange(context.Context, int64, time.Time, time.Time) ([]Event, error)
	IsBusyDateTimeRange(context.Context, int64, int64, time.Time, time.Time) error
}
