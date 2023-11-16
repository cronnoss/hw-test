package internalhttp

import (
	"context"
	"net"
	"net/http"
)

type ctxKeyID int

const (
	KeyLoggerID ctxKeyID = iota
)

type Conf struct {
	Host string `toml:"host"`
	Port string `toml:"port"`
}

type Server struct {
	srv  http.Server
	app  Application
	log  Logger
	conf Conf
}

type Logger interface {
	Fatalf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Warningf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Debugf(format string, a ...interface{})
}

type Application interface{}

func NewServer(logger Logger, app Application, conf Conf) *Server {
	return &Server{log: logger, app: app, conf: conf}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("hello-world"))
}

func (s *Server) Start(ctx context.Context) error {
	addr := net.JoinHostPort(s.conf.Host, s.conf.Port)
	midLogger := NewMiddlewareLogger()
	mux := http.NewServeMux()

	mux.Handle("/", midLogger.loggingMiddleware(http.HandlerFunc(s.ServeHTTP)))

	s.srv = http.Server{
		Addr:    addr,
		Handler: mux,
		BaseContext: func(_ net.Listener) context.Context {
			bCtx := context.WithValue(ctx, KeyLoggerID, s.log)
			return bCtx
		},
	}

	s.log.Infof("http server started on %s:%s\n", s.conf.Host, s.conf.Port)
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
