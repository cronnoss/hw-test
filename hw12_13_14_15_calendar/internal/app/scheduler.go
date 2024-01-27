package app

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/logger"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
)

type SchedulerConf struct {
	Logger  logger.Conf   `toml:"logger"`
	Storage storage.Conf  `toml:"storage"`
	URLRMQ  string        `toml:"url_rmq"`
	Period  time.Duration `toml:"period"`
}

type Scheduler struct {
	conf     SchedulerConf
	log      server.Logger
	storage  SchedulerStorage
	producer SchedulerProducer
}

type SchedulerStorage interface {
	Connect(context.Context) error
	Close(context.Context) error
	GetEventsDayOfNotice(context.Context, time.Time) ([]model.Event, error)
	DeleteEventsOlderDate(context.Context, time.Time) (int64, error)
}

type SchedulerProducer interface {
	Connect(context.Context) error
	Close(context.Context) error
	SendNotification(context.Context, *model.Event) error
}

func NewScheduler(log server.Logger, conf SchedulerConf, storage SchedulerStorage, producer SchedulerProducer,
) *Scheduler {
	scheduler := &Scheduler{
		conf:     conf,
		log:      log,
		storage:  storage,
		producer: producer,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := storage.Connect(ctx)
	if err != nil {
		server.Exitfail(fmt.Sprintf("Can't connect to storage:%v", err))
	}

	if err := producer.Connect(ctx); err != nil {
		server.Exitfail(fmt.Sprintf("Can't connect to RabbitMQ:%v", err))
	}

	return scheduler
}

func (s Scheduler) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	ticker := time.NewTicker(s.conf.Period)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			ctxStop, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelStop()
			s.Stop(ctxStop)
			return

		case <-ticker.C:
			date := time.Now()
			s.log.Debugf("Starting notification process...\n")
			sent, err := s.SendNotification(ctx, date)
			if err != nil {
				s.log.Errorf("%v", err)
				return
			}
			s.log.Debugf("Notifications  sent:%v\n", sent)

			s.log.Debugf("Starting to remove events that are older than a year\n")
			deleted, err := s.DeleteEventsOlderDate(ctx, date.AddDate(-1, 0, 0))
			if err != nil {
				s.log.Errorf("%v", err)
				return
			}
			s.log.Debugf("Old events deleted:%v\n", deleted)
			s.log.Debugf("Notification process has finished\n")
		}
	}
}

func (s Scheduler) Stop(ctx context.Context) {
	s.producer.Close(ctx)
	s.log.Debugf("Producer closed\n")
	s.storage.Close(ctx)
	s.log.Debugf("Storage closed\n")
}

func (s Scheduler) DeleteEventsOlderDate(ctx context.Context, date time.Time) (int64, error) {
	return s.storage.DeleteEventsOlderDate(ctx, date)
}

func (s Scheduler) SendNotification(ctx context.Context, date time.Time) (int64, error) {
	sent := int64(0)
	events, err := s.storage.GetEventsDayOfNotice(ctx, date)
	if err != nil {
		return sent, err
	}

	for i := range events {
		err := s.producer.SendNotification(ctx, &events[i])
		if err != nil {
			return sent, fmt.Errorf("SendNotification:%w", err)
		}
		sent++
	}
	return sent, nil
}
