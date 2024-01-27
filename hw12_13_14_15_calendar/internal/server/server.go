package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
)

var (
	ErrID             = errors.New("wrong ID")
	ErrUserID         = errors.New("wrong UserID")
	ErrTitle          = errors.New("wrong Title")
	ErrDescription    = errors.New("wrong Description")
	ErrOnTime         = errors.New("wrong OnTime")
	ErrOffTime        = errors.New("wrong OffTime")
	ErrNotifyTime     = errors.New("wrong NotifyTime")
	ErrEventNotFound  = errors.New("event not found")
	ErrTooLongCloseDB = errors.New("too long close db")
)

//go:generate mockery --name Logger
type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

//go:generate mockery --name Application
type Application interface {
	InsertEvent(context.Context, *model.Event) error
	UpdateEvent(context.Context, *model.Event) error
	DeleteEvent(context.Context, int64) error
	GetEventByID(context.Context, int64) (model.Event, error)
	GetAllEvents(context.Context, int64) ([]model.Event, error)
	GetAllEventsDay(context.Context, int64, time.Time) ([]model.Event, error)
	GetAllEventsWeek(context.Context, int64, time.Time) ([]model.Event, error)
	GetAllEventsMonth(context.Context, int64, time.Time) ([]model.Event, error)
}

func Exitfail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
