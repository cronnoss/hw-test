//go:generate protoc --go_out=../../pkg/event_service_v1 --proto_path=../../api/ ../../api/EventService.proto
//go:generate protoc --go-grpc_out=../../pkg/event_service_v1 --proto_path=../../api/ ../../api/EventService.proto
package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/app"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/logger"
	internalgrpc "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server/grpc"
	internalhttp "github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server/http"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
)

func main() {
	conf := NewConfig().CalendarConf
	storage := storage.NewStorage(conf.Storage)
	logger := logger.NewLogger(conf.Logger.Level, os.Stdout)
	calendar := app.NewCalendar(logger, conf, storage)
	httpsrv := internalhttp.NewServer(logger, calendar, conf.HTTP.Host, conf.HTTP.Port)
	grpcsrv, _ := internalgrpc.NewServer(logger, calendar, conf.GRPC.Host, conf.GRPC.Port)

	calendar.Run(httpsrv, grpcsrv)

	filename := filepath.Base(os.Args[0])
	fmt.Printf("%s stopped\n", filename)
}
