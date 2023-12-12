package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
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
	ErrDateBusy       = errors.New("date is busy")
	ErrTooLongCloseDB = errors.New("too long close db")
)

type App struct {
	log     Logger
	storage storage.Storage
}

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

func New(logger Logger, storage storage.Storage) *App {
	return &App{log: logger, storage: storage}
}

func CheckingEvent(e *storage.Event, checkID bool) error {
	if checkID && e.ID == 0 {
		return fmt.Errorf("%w(ID is zero)", ErrID)
	}
	if e.UserID == 0 {
		return fmt.Errorf("%w(UserID is %v)", ErrUserID, e.UserID)
	}

	if len(e.Title) > 150 {
		return fmt.Errorf("%w(len %v, must be <=150)", ErrTitle, len(e.Title))
	}

	if e.OnTime.IsZero() {
		return fmt.Errorf("%w(empty OnTime)", ErrOnTime)
	}

	switch {
	case e.OffTime.IsZero():
		return fmt.Errorf("%w: empty", ErrOffTime)
	case e.OffTime.Before(e.OnTime):
		return fmt.Errorf("%w: before OnTime", ErrOffTime)
	case e.OffTime.Equal(e.OnTime):
		return fmt.Errorf("%w: equal OnTime", ErrOffTime)
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

func (a *App) InsertEvent(ctx context.Context, event *storage.Event) error {
	if err := CheckingEvent(event, false); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.InsertEvent(ctx, event)
}

func (a *App) UpdateEvent(ctx context.Context, event *storage.Event) error {
	if err := CheckingEvent(event, true); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.UpdateEvent(ctx, event)
}

func (a *App) DeleteEvent(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.DeleteEvent(ctx, id)
}

func (a *App) GetEventByID(ctx context.Context, id int64) (storage.Event, error) {
	if id == 0 {
		return storage.Event{}, ErrID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetEventByID(ctx, id)
}

func (a *App) GetAllEvents(ctx context.Context, userID int64) ([]storage.Event, error) {
	if userID == 0 {
		return []storage.Event{}, ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetAllEvents(ctx, userID)
}

func (a *App) GetAllEventsDay(ctx context.Context, userID int64, date time.Time) ([]storage.Event, error) {
	if userID == 0 {
		return []storage.Event{}, ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetAllRange(ctx, userID, date, date.AddDate(0, 0, 1))
}

func (a *App) GetAllEventsWeek(ctx context.Context, userID int64, date time.Time) ([]storage.Event, error) {
	if userID == 0 {
		return []storage.Event{}, ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetAllRange(ctx, userID, date, date.AddDate(0, 0, 7))
}

func (a *App) GetAllEventsMonth(ctx context.Context, userID int64, date time.Time) ([]storage.Event, error) {
	if userID == 0 {
		return []storage.Event{}, ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetAllRange(ctx, userID, date, date.AddDate(0, 1, 0))
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello-world"))
}
