package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage/sql"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	var storage app.Storage

	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()
	if err := config.LoadConfigFile(configFile); err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("Config:", config)

	log, err := logger.New(config.Logger.Level, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Can't allocate logger:%v\n", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer cancel()

	switch config.Storage.DB {
	case "in-memory":
		db := memorystorage.New()
		storage = db
	case "sql":
		db := sqlstorage.New(config.Storage.DSN)
		err := db.Connect(ctx)
		if err != nil {
			log.Errorf("failed to connect to db: " + err.Error())
		}
		storage = db
	}

	calendar := app.New(log, storage)
	server := internalhttp.NewServer(log, calendar, config.HTTPServer)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := server.Stop(ctx); err != nil {
			log.Errorf("failed to stop http server: " + err.Error())
		}
	}()

	log.Infof("calendar is running...\n")

	if err := server.Start(ctx); err != nil {
		log.Errorf("failed to start http server: " + err.Error())
		cancel()
		os.Exit(1) //nolint:gocritic
	}
}
