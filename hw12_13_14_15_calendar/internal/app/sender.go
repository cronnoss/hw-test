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

type SenderConf struct {
	Logger  logger.Conf  `toml:"logger"`
	Storage storage.Conf `toml:"storage"`
	URLRMQ  string       `toml:"url_rmq"`
}

type Sender struct {
	conf     SenderConf
	log      server.Logger
	storage  SenderStorage
	consumer SenderConsumer
}

type SenderStorage interface {
	Connect(context.Context) error
	Close(context.Context) error

	UpdateEventNotified(context.Context, int64) error
}

type SenderConsumer interface {
	Connect(context.Context) error
	Close(context.Context) error
	NotifyChannel() <-chan model.NotificationMsg
}

func NewSender(log server.Logger, conf SenderConf, storage SenderStorage, consumer SenderConsumer) *Sender {
	sender := &Sender{
		conf:     conf,
		log:      log,
		storage:  storage,
		consumer: consumer,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := storage.Connect(ctx)
	if err != nil {
		server.Exitfail(fmt.Sprintf("Can't connect to storage:%v", err))
	}

	if err := consumer.Connect(ctx); err != nil {
		server.Exitfail(fmt.Sprintf("Can't connect to RabbitMQ:%v", err))
	}

	return sender
}

func (s Sender) Run() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			ctxStop, cancelStop := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancelStop()
			s.Stop(ctxStop)
			return

		case msg, ok := <-s.consumer.NotifyChannel():
			if ok {
				err := s.storage.UpdateEventNotified(ctx, msg.ID)
				if err != nil {
					s.log.Errorf("Can't update notified:%v\n", err)
					continue
				}
				s.log.Debugf("UpdateEventNotified: updated\n")
			}
		}
	}
}

func (s Sender) Stop(ctx context.Context) {
	s.consumer.Close(ctx)
	s.log.Debugf("Consumer closed\n")
	s.storage.Close(ctx)
	s.log.Debugf("Storage closed\n")
}
