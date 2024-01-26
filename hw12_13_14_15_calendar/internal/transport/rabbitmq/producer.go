package internalrmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

type Producer struct {
	log     Logger
	url     string
	channel *amqp.Channel
	connect *amqp.Connection
	queue   amqp.Queue
}

var ErrCantSendMsg = errors.New("can't send message")

func NewProducer(log Logger, url string) *Producer {
	return &Producer{log: log, url: url}
}

func (c *Producer) Connect(ctx context.Context) error {
	var err error
	c.log.Debugf("Connecting to RabbitMQ...\n")
	c.connect, err = amqp.Dial(c.url)
	if err != nil {
		return err
	}

	c.channel, err = c.connect.Channel()
	if err != nil {
		return err
	}

	c.queue, err = c.channel.QueueDeclare(getQueueDeclated())
	if err != nil {
		return err
	}

	c.log.Debugf("Connected to RabbitMQ\n")

	return nil
}

func (c *Producer) Close(ctx context.Context) error {
	c.connect.Close()
	c.channel.Close()
	return nil
}

func (c *Producer) Start(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (c *Producer) Stop(ctx context.Context) error {
	return nil
}

func (c *Producer) SendNotification(ctx context.Context, event *model.Event) error {
	msg := model.NotificationMsg{
		ID:     event.ID,
		Title:  event.Title,
		Date:   event.OnTime,
		UserID: event.UserID,
	}

	jdata, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	pub := amqp.Publishing{
		ContentType: "application/json",
		Body:        jdata,
	}

	if err := c.channel.PublishWithContext(ctx, "", c.queue.Name, false, false, pub); err != nil {
		return fmt.Errorf("SendNotification: %w", err)
	}
	c.log.Debugf("sent notification msg\n")
	return nil
}

func getQueueDeclated() (string, bool, bool, bool, bool, amqp.Table) {
	return "notification", false, false, false, false, nil
}
