package server

import (
	"context"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
)

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

type Application interface {
	InsertEvent(context.Context, *storage.Event) error
	UpdateEvent(context.Context, *storage.Event) error
	DeleteEvent(context.Context, int64) error
	GetEventByID(context.Context, int64) (storage.Event, error)
	GetAllEvents(context.Context, int64) ([]storage.Event, error)
	GetAllEventsDay(context.Context, int64, time.Time) ([]storage.Event, error)
	GetAllEventsWeek(context.Context, int64, time.Time) ([]storage.Event, error)
	GetAllEventsMonth(context.Context, int64, time.Time) ([]storage.Event, error)
}
