package internalhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/model"
	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server"
)

type ctxKeyID int

const (
	KeyLoggerID ctxKeyID = iota
)

type Server struct {
	srv  http.Server
	app  server.Application
	log  Logger
	host string
	port string
}

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

type reqByID struct {
	ID int64 `json:"id"`
}

type reqByUser struct {
	UserID int64 `json:"userid"`
}

type reqByUserByDate struct {
	UserID int64     `json:"userid"`
	Date   time.Time `json:"date"`
}

func NewServer(log Logger, app server.Application, host, port string) *Server {
	return &Server{log: log, app: app, host: host, port: port}
}

func (s *Server) helperDecode(stream io.ReadCloser, w http.ResponseWriter, data interface{}) error {
	decoder := json.NewDecoder(stream)
	if err := decoder.Decode(&data); err != nil {
		s.log.Errorf("Can't decode json:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't decode json:%v\"}\n", err)))
		return err
	}
	return nil
}

func (s *Server) InsertEvent(w http.ResponseWriter, r *http.Request) {
	var event model.Event
	if err := s.helperDecode(r.Body, w, &event); err != nil {
		return
	}

	err := s.app.InsertEvent(r.Context(), &event)
	if err != nil {
		s.log.Errorf("InsertEvent:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't InsertEvent:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"msg\": \"Inserted\"}\n"))
}

func (s *Server) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var event model.Event
	if err := s.helperDecode(r.Body, w, &event); err != nil {
		return
	}
	if err := s.app.UpdateEvent(r.Context(), &event); err != nil {
		s.log.Errorf("Can't update event:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't update event:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"msg\": \"Updated\"}\n"))
}

func (s *Server) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	var req reqByID
	if err := s.helperDecode(r.Body, w, &req); err != nil {
		return
	}
	if err := s.app.DeleteEvent(r.Context(), req.ID); err != nil {
		s.log.Errorf("Can't delete event:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't delete event:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"msg\": \"Deleted\"}\n"))
}

func (s *Server) GetEventByID(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	var req reqByID
	if err := s.helperDecode(r.Body, w, &req); err != nil {
		return
	}
	event, err := s.app.GetEventByID(r.Context(), req.ID)
	if err != nil {
		s.log.Errorf("Can't get event by id:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get event by id:%v\"}\n", err)))
		return
	}
	rawJSON, err := json.Marshal(event)
	if err != nil {
		s.log.Errorf("Can't marshal event:%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't marshal event:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(rawJSON)
}

func (s *Server) GetAllEvents(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	var req reqByUser
	if err := s.helperDecode(r.Body, w, &req); err != nil {
		return
	}
	events, err := s.app.GetAllEvents(r.Context(), req.UserID)
	if err != nil {
		s.log.Errorf("Can't get all events:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get all events:%v\"}\n", err)))
		return
	}
	rawJSON, err := json.Marshal(events)
	if err != nil {
		s.log.Errorf("Can't marshal events:%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't marshal events:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(rawJSON)
}

func (s *Server) GetAllEventsDay(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	var req reqByUserByDate
	if err := s.helperDecode(r.Body, w, &req); err != nil {
		return
	}
	events, err := s.app.GetAllEventsDay(r.Context(), req.UserID, req.Date)
	if err != nil {
		s.log.Errorf("Can't get all events:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get all events:%v\"}\n", err)))
		return
	}
	rawJSON, err := json.Marshal(events)
	if err != nil {
		s.log.Errorf("Can't marshal events:%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't marshal events:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(rawJSON)
}

func (s *Server) GetAllEventsWeek(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	var req reqByUserByDate
	if err := s.helperDecode(r.Body, w, &req); err != nil {
		return
	}
	events, err := s.app.GetAllEventsWeek(r.Context(), req.UserID, req.Date)
	if err != nil {
		s.log.Errorf("Can't get all events:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get all events:%v\"}\n", err)))
		return
	}
	rawJSON, err := json.Marshal(events)
	if err != nil {
		s.log.Errorf("Can't marshal events:%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't marshal events:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(rawJSON)
}

func (s *Server) GetAllEventsMonth(w http.ResponseWriter, r *http.Request) { //nolint:dupl
	var req reqByUserByDate
	if err := s.helperDecode(r.Body, w, &req); err != nil {
		return
	}
	events, err := s.app.GetAllEventsMonth(r.Context(), req.UserID, req.Date)
	if err != nil {
		s.log.Errorf("Can't get all events:%v\n", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't get all events:%v\"}\n", err)))
		return
	}
	rawJSON, err := json.Marshal(events)
	if err != nil {
		s.log.Errorf("Can't marshal events:%v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("{\"error\": \"Can't marshal events:%v\"}\n", err)))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(rawJSON)
}

func (s *Server) Start(ctx context.Context) error {
	addr := net.JoinHostPort(s.host, s.port)
	midLogger := NewMiddlewareLogger()
	mux := http.NewServeMux()

	mux.Handle("/InsertEvent", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.InsertEvent))))
	mux.Handle("/UpdateEvent", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.UpdateEvent))))
	mux.Handle("/DeleteEvent", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.DeleteEvent))))
	mux.Handle("/GetEventByID", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.GetEventByID))))
	mux.Handle("/GetAllEvents", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.GetAllEvents))))
	mux.Handle("/GetAllEventsDay", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.GetAllEventsDay))))
	mux.Handle("/GetAllEventsWeek", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.GetAllEventsWeek))))
	mux.Handle("/GetAllEventsMonth", midLogger.setCommonHeadersMiddleware(
		midLogger.loggingMiddleware(http.HandlerFunc(s.GetAllEventsMonth))))

	s.srv = http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			bCtx := context.WithValue(ctx, KeyLoggerID, s.log)
			return bCtx
		},
	}

	s.log.Infof("http server started on %s:%s\n", s.host, s.port)
	if err := s.srv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	if err := s.srv.Shutdown(ctx); err != nil {
		return err
	}
	s.log.Infof("http server shutdown\n")
	return nil
}
