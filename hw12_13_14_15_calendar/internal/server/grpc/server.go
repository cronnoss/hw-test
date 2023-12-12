package grpc

import (
	"context"
	"net"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/storage"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/pkg/event_service_v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ctxKeyID int

const (
	KeyMethodID ctxKeyID = iota
)

type Conf struct {
	Port string `toml:"port"`
	Host string `toml:"host"`
}

type Server struct {
	event_service_v1.UnimplementedEventServiceV1Server

	srv  *grpc.Server
	app  server.Application
	log  server.Logger
	conf Conf
}

func (Server) APIEventFromEvent(event *storage.Event) *event_service_v1.Event {
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

func (Server) EventFromAPIEvent(apiEvent *event_service_v1.Event) *storage.Event {
	event := storage.Event{}

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

func NewServer(log server.Logger, app server.Application, conf Conf, srv *grpc.Server) *Server {
	return &Server{
		app:  app,
		conf: conf,
		log:  log,
		srv:  srv,
	}
}

func (s *Server) Start(context.Context) error {
	addr := net.JoinHostPort(s.conf.Host, s.conf.Port)
	dial, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	event_service_v1.RegisterEventServiceV1Server(s.srv, s)

	s.log.Infof("grpc server started on %v\n", addr)
	if err := s.srv.Serve(dial); err != nil {
		return err
	}

	return nil
}

func (s *Server) Stop(context.Context) error {
	s.srv.GracefulStop()
	s.log.Infof("grpc server shutdown\n")
	return nil
}
