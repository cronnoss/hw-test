//go:generate protoc --go_out=../../pkg/event_service_v1 --proto_path=../../api/ ../../api/EventService.proto
//go:generate protoc --go-grpc_out=../../pkg/event_service_v1 --proto_path=../../api/ ../../api/EventService.proto
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server/http"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage/memory"
	sqlstorage "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage/sql"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "/etc/calendar/config.toml", "Path to configuration file")
}

func main() {
	var appStorage storage.Storage

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
	case "in_memory":
		db := memorystorage.New()
		appStorage = db
	case "sql":
		db := sqlstorage.New(config.Storage.DSN)
		err := db.Connect(ctx)
		if err != nil {
			log.Errorf("failed to connect to db: " + err.Error())
		}
		appStorage = db
	}

	calendar := app.New(log, appStorage)
	httpServer := internalhttp.NewServer(log, calendar, config.HTTPServer, cancel)

	unarayLoggerEnricherIntercepter := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		timeStart := time.Now()
		var b strings.Builder
		ip, _ := peer.FromContext(ctx)
		md, ok := metadata.FromIncomingContext(ctx)
		userAgent := "unknown"

		if ok {
			userAgent = md["user-agent"][0]
		}

		b.WriteString(ip.Addr.String())
		b.WriteString(" ")
		b.WriteString(timeStart.Format("[02/Jan/2006:15:04:05 -0700]"))
		b.WriteString(" ")
		b.WriteString(info.FullMethod)
		b.WriteString(" ")
		b.WriteString(time.Since(timeStart).String())
		b.WriteString(" ")
		b.WriteString(userAgent)
		b.WriteString("\n")
		log.Infof(b.String())
		return handler(ctx, req)
	}

	srv := grpc.NewServer(grpc.UnaryInterceptor(unarayLoggerEnricherIntercepter))
	grpcServer := internalgrpc.NewServer(log, calendar, config.GRPSServer, srv)

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		if err := httpServer.Stop(ctx); err != nil {
			if !errors.Is(err, http.ErrServerClosed) &&
				!errors.Is(err, context.Canceled) {
				log.Errorf("failed to stop HTTP-httpServer:%v\n", err)
			}
		}

		if err := grpcServer.Stop(ctx); err != nil {
			if !errors.Is(err, grpc.ErrServerStopped) {
				log.Errorf("failed to stop GRPC-httpServer:%v\n", err)
			}
		}

		if err := appStorage.Close(ctx); err != nil {
			log.Errorf("failed to close storage:%v\n", err)
		}
	}()

	log.Infof("calendar is running...\n")

	g, ctxEG := errgroup.WithContext(ctx)
	func1 := func() error {
		return httpServer.Start(ctxEG)
	}

	func2 := func() error {
		return grpcServer.Start(ctxEG)
	}

	g.Go(func1)
	g.Go(func2)

	if err := g.Wait(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) &&
			!errors.Is(err, grpc.ErrServerStopped) &&
			!errors.Is(err, context.Canceled) {
			log.Errorf("%v\n", err)
		}
	}
}
