package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/logger"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
	internalrmq "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/transport/rabbitmq"
)

func main() {
	conf := NewConfig().SchedulerConf
	storage := storage.NewStorage(conf.Storage)
	logger := logger.NewLogger(conf.Logger.Level, os.Stdout)
	producer := internalrmq.NewProducer(logger, conf.URLRMQ)
	scheduler := app.NewScheduler(logger, conf, storage, producer)

	scheduler.Run()

	filename := filepath.Base(os.Args[0])
	fmt.Printf("%s stopped\n", filename)
}
