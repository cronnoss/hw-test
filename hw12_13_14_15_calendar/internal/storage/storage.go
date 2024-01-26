package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	memorystorage "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage/sql"
)

type Conf struct {
	DB  string `toml:"db"`
	DSN string `toml:"dsn"`
}

type Storage interface {
	Connect(context.Context) error
	Close(context.Context) error
	InsertEvent(context.Context, *model.Event) error
	UpdateEvent(context.Context, *model.Event) error
	DeleteEvent(context.Context, int64) error
	GetEventByID(context.Context, int64) (model.Event, error)
	GetAllEvents(context.Context, int64) ([]model.Event, error)
	GetAllRange(context.Context, int64, time.Time, time.Time) ([]model.Event, error)
	IsBusyDateTimeRange(context.Context, int64, int64, time.Time, time.Time) error

	// for producers
	GetEventsDayOfNotice(context.Context, time.Time) ([]model.Event, error)
	DeleteEventsOlderDate(context.Context, time.Time) (int64, error)

	// for consumers
	UpdateEventNotified(context.Context, int64) error
}

func NewStorage(conf Conf) Storage {
	switch conf.DB {
	case "in_memory":
		return memorystorage.New()
	case "sql":
		return sqlstorage.New(conf.DSN)
	}

	fmt.Fprintln(os.Stderr, "wrong DB")
	os.Exit(1)
	return nil
}
