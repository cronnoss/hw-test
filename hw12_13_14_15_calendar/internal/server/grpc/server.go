package internalgrpc

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/pkg/event_service_v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ctxKeyID int

const (
	KeyMethodID ctxKeyID = iota
)

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

type Server struct {
	event_service_v1.UnimplementedEventServiceV1Server

	basesrv *grpc.Server
	app     server.Application
	log     Logger
	host    string
	port    string
}

func (Server) APIEventFromEvent(event *model.Event) *event_service_v1.Event {
	return &event_service_v1.Event{
		ID:          &event.ID,
		UserID:      &event.UserID,
		Title:       &event.Title,
		Description: &event.Description,
		OnTime:      timestamppb.New(event.OnTime),
		OffTime:     timestamppb.New(event.OffTime),
		NotifyTime:  timestamppb.New(event.NotifyTime),
	}
}

func (Server) EventFromAPIEvent(apiEvent *event_service_v1.Event) *model.Event {
	event := model.Event{}

	event.ID = *apiEvent.ID
	event.UserID = *apiEvent.UserID
	event.Title = *apiEvent.Title
	event.Description = *apiEvent.Description
	if err := apiEvent.OnTime.CheckValid(); err == nil {
		event.OnTime = apiEvent.OnTime.AsTime().Local()
	}
	if err := apiEvent.OffTime.CheckValid(); err == nil {
		event.OffTime = apiEvent.OffTime.AsTime().Local()
	}
	if err := apiEvent.NotifyTime.CheckValid(); err == nil {
		event.NotifyTime = apiEvent.NotifyTime.AsTime().Local()
	}

	return &event
}

func (s *Server) InsertEvent(ctx context.Context, req *event_service_v1.ReqByEvent) (*event_service_v1.RepID, error) {
	event := s.EventFromAPIEvent(req.Event)
	if err := s.app.InsertEvent(ctx, event); err != nil {
		return nil, err
	}
	return &event_service_v1.RepID{ID: &event.ID}, nil
}

func (s Server) UpdateEvent(ctx context.Context, req *event_service_v1.ReqByEvent) (*emptypb.Empty, error) {
	event := s.EventFromAPIEvent(req.Event)
	if err := s.app.UpdateEvent(ctx, event); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s Server) DeleteEvent(ctx context.Context, req *event_service_v1.ReqByID) (*emptypb.Empty, error) {
	if err := s.app.DeleteEvent(ctx, *req.ID); err != nil {
		return nil, err
	}
	return new(emptypb.Empty), nil
}

func (s *Server) GetEventByID(ctx context.Context, req *event_service_v1.ReqByID) (*event_service_v1.RepEvents, error) {
	event, err := s.app.GetEventByID(ctx, *req.ID)
	_ = event
	if err != nil {
		return nil, err
	}

	rep := event_service_v1.RepEvents{}
	rep.Event = append(rep.Event, s.APIEventFromEvent(&event))
	return &rep, nil
}

func (s *Server) GetAllEvents(
	ctx context.Context,
	req *event_service_v1.ReqByUser,
) (*event_service_v1.RepEvents, error) {
	events, err := s.app.GetAllEvents(ctx, *req.UserID)
	if err != nil {
		return nil, err
	}

	rep := event_service_v1.RepEvents{}
	rep.Event = make([]*event_service_v1.Event, len(events))
	for i, event := range events {
		event := event
		rep.Event[i] = s.APIEventFromEvent(&event)
	}
	return &rep, nil
}

func (s *Server) GetAllEventsDay(
	ctx context.Context,
	req *event_service_v1.ReqByUserByDate,
) (*event_service_v1.RepEvents, error) {
	events, err := s.app.GetAllEventsDay(ctx, *req.UserID, req.Date.AsTime().Local())
	if err != nil {
		return nil, err
	}

	rep := event_service_v1.RepEvents{}
	rep.Event = make([]*event_service_v1.Event, len(events))
	for i, event := range events {
		event := event
		rep.Event[i] = s.APIEventFromEvent(&event)
	}
	return &rep, nil
}

func (s *Server) GetAllEventsWeek(
	ctx context.Context,
	req *event_service_v1.ReqByUserByDate,
) (*event_service_v1.RepEvents, error) {
	events, err := s.app.GetAllEventsWeek(ctx, *req.UserID, req.Date.AsTime().Local())
	if err != nil {
		return nil, err
	}

	rep := event_service_v1.RepEvents{}
	rep.Event = make([]*event_service_v1.Event, len(events))
	for i, event := range events {
		event := event
		rep.Event[i] = s.APIEventFromEvent(&event)
	}
	return &rep, nil
}

func (s *Server) GetAllEventsMonth(
	ctx context.Context,
	req *event_service_v1.ReqByUserByDate,
) (*event_service_v1.RepEvents, error) {
	events, err := s.app.GetAllEventsMonth(ctx, *req.UserID, req.Date.AsTime().Local())
	if err != nil {
		return nil, err
	}

	rep := event_service_v1.RepEvents{}
	rep.Event = make([]*event_service_v1.Event, len(events))
	for i, event := range events {
		event := event
		rep.Event[i] = s.APIEventFromEvent(&event)
	}
	return &rep, nil
}

func NewServer(log Logger, app server.Application, host, port string) (*Server, *grpc.Server) {
	unarayLoggerEnricherIntercepter := func(ctx context.Context,
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

	basesrv := grpc.NewServer(grpc.UnaryInterceptor(unarayLoggerEnricherIntercepter))

	serverGrpc := &Server{
		log:                               log,
		app:                               app,
		basesrv:                           basesrv,
		host:                              host,
		port:                              port,
		UnimplementedEventServiceV1Server: event_service_v1.UnimplementedEventServiceV1Server{},
	}

	event_service_v1.RegisterEventServiceV1Server(serverGrpc.basesrv, serverGrpc)

	return serverGrpc, basesrv
}

func (s *Server) Start(context.Context) error {
	addr := net.JoinHostPort(s.host, s.port)
	dial, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.log.Infof("grpc server started on %v\n", addr)
	if err := s.basesrv.Serve(dial); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop(context.Context) error {
	s.basesrv.GracefulStop()
	s.log.Infof("grpc server shutdown\n")
	return nil
}
