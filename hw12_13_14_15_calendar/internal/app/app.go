package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
)

var (
	ErrUserID         = errors.New("wrong UserID")
	ErrTitle          = errors.New("wrong Title")
	ErrDescription    = errors.New("wrong Description")
	ErrOnTime         = errors.New("wrong OnTime")
	ErrOffTime        = errors.New("wrong OffTime")
	ErrNotifyTime     = errors.New("wrong NotifyTime")
	ErrEventNotFound  = errors.New("event not found")
	ErrDateBusy       = errors.New("date is busy")
	ErrTooLongCloseDB = errors.New("too long close db")
)

type App struct {
	log     Logger
	storage Storage
}

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

type Storage interface {
	Connect(context.Context) error
	Close(context.Context) error
	InsertEvent(context.Context, *storage.Event) error
	UpdateEvent(context.Context, *storage.Event) error
	DeleteEvent(context.Context, *storage.Event) error
	GetEventByID(ctx context.Context, eID int64) (storage.Event, error)
	GetAll(ctx context.Context, userID int64) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{log: logger, storage: storage}
}

func CheckingEvent(e *storage.Event) error {
	if e.UserID == 0 {
		return fmt.Errorf("%w(UserID is %v)", ErrUserID, e.UserID)
	}

	if len(e.Title) > 150 {
		return fmt.Errorf("%w(len %v, must be <=150)", ErrTitle, len(e.Title))
	}

	if e.OnTime.IsZero() {
		return fmt.Errorf("%w(empty OnTime)", ErrOnTime)
	}

	if !e.OffTime.IsZero() {
		if e.OffTime.Before(e.OnTime) {
			return fmt.Errorf("%w(OffTime before OnTime)", ErrOffTime)
		}
		if e.OffTime.Equal(e.OnTime) {
			return fmt.Errorf("%w(OffTime equal OnTime)", ErrOffTime)
		}
	}

	if !e.NotifyTime.IsZero() {
		if e.NotifyTime.After(e.OffTime) {
			return fmt.Errorf("%w(NotifyTime after OffTime)", ErrNotifyTime)
		}
		if e.NotifyTime.Before(e.OnTime) {
			return fmt.Errorf("%w(NotifyTime before OnTime)", ErrNotifyTime)
		}
	}

	return nil
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	// TODO
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}

func (m *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello-world"))
}

// TODO
