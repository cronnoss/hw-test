package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/logger"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type CalendarConf struct {
	Logger  logger.Conf  `toml:"logger"`
	Storage storage.Conf `toml:"storage"`
	HTTP    struct {
		Host string `toml:"host"`
		Port string `toml:"port"`
	} `toml:"http-server"`
	GRPC struct {
		Host string `toml:"host"`
		Port string `toml:"port"`
	} `toml:"grpc-server"`
}

type Calendar struct {
	conf    CalendarConf
	log     server.Logger
	storage Storage
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
}

type Server interface {
	Start(context.Context) error
	Stop(context.Context) error
}

func (a *Calendar) CheckingEvent(e *model.Event, checkID bool) error {
	if checkID && e.ID == 0 {
		return fmt.Errorf("%w(ID is zero)", server.ErrID)
	}
	if e.UserID == 0 {
		return fmt.Errorf("%w(UserID is %v)", server.ErrUserID, e.UserID)
	}

	if len(e.Title) > 150 {
		return fmt.Errorf("%w(len %v, must be <=150)", server.ErrTitle, len(e.Title))
	}

	if e.OnTime.IsZero() {
		return fmt.Errorf("%w(empty OnTime)", server.ErrOnTime)
	}

	switch {
	case e.OffTime.IsZero():
		return fmt.Errorf("%w: empty", server.ErrOffTime)
	case e.OffTime.Before(e.OnTime):
		return fmt.Errorf("%w: before OnTime", server.ErrOffTime)
	case e.OffTime.Equal(e.OnTime):
		return fmt.Errorf("%w: equal OnTime", server.ErrOffTime)
	}

	if !e.NotifyTime.IsZero() {
		if e.NotifyTime.After(e.OffTime) {
			return fmt.Errorf("%w(NotifyTime after OffTime)", server.ErrNotifyTime)
		}
	}

	return nil
}

func (a *Calendar) isBusyDateTimeRange(ctx context.Context, id, userID int64, onTime, offTime time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.IsBusyDateTimeRange(ctx, id, userID, onTime, offTime)
}

func (a *Calendar) firstDayOfWeek(t time.Time) time.Time {
	for t.Weekday() != time.Monday {
		t = t.AddDate(0, 0, -1)
	}
	return t
}

func (a *Calendar) firstDayOfMonth(t time.Time) time.Time {
	return t.AddDate(0, 0, -t.Day()+1)
}

func (a *Calendar) lastDayOfMonth(t time.Time) time.Time {
	return t.AddDate(0, 1, -t.Day())
}

func (a *Calendar) Close(ctx context.Context) error {
	a.log.Infof("App closed\n")
	return a.storage.Close(ctx)
}

func (a *Calendar) InsertEvent(ctx context.Context, event *model.Event) error {
	if err := a.CheckingEvent(event, false); err != nil {
		return err
	}

	if err := a.isBusyDateTimeRange(ctx, event.ID, event.UserID, event.OnTime, event.OffTime); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return a.storage.InsertEvent(ctx, event)
}

func (a *Calendar) UpdateEvent(ctx context.Context, event *model.Event) error {
	if err := a.CheckingEvent(event, true); err != nil {
		return err
	}

	if err := a.isBusyDateTimeRange(ctx, event.ID, event.UserID, event.OnTime, event.OffTime); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.UpdateEvent(ctx, event)
}

func (a *Calendar) DeleteEvent(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.DeleteEvent(ctx, id)
}

func (a *Calendar) GetEventByID(ctx context.Context, id int64) (model.Event, error) {
	if id == 0 {
		return model.Event{}, server.ErrID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetEventByID(ctx, id)
}

func (a *Calendar) GetAllEvents(ctx context.Context, userID int64) ([]model.Event, error) {
	if userID == 0 {
		return []model.Event{}, server.ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetAllEvents(ctx, userID)
}

func (a *Calendar) GetAllEventsDay(ctx context.Context, userID int64, date time.Time) ([]model.Event, error) {
	if userID == 0 {
		return []model.Event{}, server.ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return a.storage.GetAllRange(ctx, userID, date, date)
}

func (a *Calendar) GetAllEventsWeek(ctx context.Context, userID int64, date time.Time) ([]model.Event, error) {
	if userID == 0 {
		return []model.Event{}, server.ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	monday := a.firstDayOfWeek(date)
	sunday := monday.AddDate(0, 0, 6)
	return a.storage.GetAllRange(ctx, userID, monday, sunday)
}

func (a *Calendar) GetAllEventsMonth(ctx context.Context, userID int64, date time.Time) ([]model.Event, error) {
	if userID == 0 {
		return []model.Event{}, server.ErrUserID
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	dayFirst := a.firstDayOfMonth(date)
	dayLast := a.lastDayOfMonth(date)
	return a.storage.GetAllRange(ctx, userID, dayFirst, dayLast)
}

func NewCalendar(log server.Logger, conf CalendarConf, storage Storage) *Calendar {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := storage.Connect(ctx); err != nil {
		server.Exitfail(fmt.Sprintf("Can't connect to storage:%v", err))
	}

	return &Calendar{log: log, conf: conf, storage: storage}
}

func (a Calendar) Run(httpsrv Server, grpcsrv Server) {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	g, ctxEG := errgroup.WithContext(ctx)

	func1 := func() error {
		return httpsrv.Start(ctxEG)
	}

	func2 := func() error {
		return grpcsrv.Start(ctxEG)
	}

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := grpcsrv.Stop(ctx); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				a.log.Errorf("failed to stop GRPC-server:%v\n", err)
			}
		}

		if err := httpsrv.Stop(ctx); err != nil {
			if !errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, context.Canceled) {
				a.log.Errorf("failed to stop HTTP-server:%v\n", err)
			}
		}

		if err := a.storage.Close(ctx); err != nil {
			a.log.Errorf("failed to close db:%v\n", err)
		}
	}()

	g.Go(func1)
	g.Go(func2)

	if err := g.Wait(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) &&
			!errors.Is(err, grpc.ErrServerStopped) &&
			!errors.Is(err, context.Canceled) {
			a.log.Errorf("%v\n", err)
		}
	}
}
