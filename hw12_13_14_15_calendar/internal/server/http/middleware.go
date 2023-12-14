package internalhttp

import (
	"net/http"
	"time"

	"github.com/cronnoss/hw-test/hw12_13_14_15_calendar/internal/server"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

type MiddlewareLogger struct{}

func NewMiddlewareLogger() *MiddlewareLogger {
	return &MiddlewareLogger{}
}

func (a *MiddlewareLogger) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := &statusWriter{ResponseWriter: w}

		l := r.Context().Value(KeyLoggerID).(server.Logger)
		start := time.Now()

		next.ServeHTTP(sw, r)

		l.Debugf("%s [%s] %s %s %s %d %s %s %s\n",
			r.RemoteAddr,
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.RequestURI,
			r.Proto,
			sw.status,
			http.StatusText(sw.status),
			time.Since(start).String(),
			r.Header.Get("User-Agent"),
		)
	})
}

func (a *MiddlewareLogger) setCommonHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
