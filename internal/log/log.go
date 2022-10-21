package log

import (
	"net/http"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/go-chi/chi/v5/middleware"
)

// Code in this package matches work from @negz in
// https://github.com/upbound/xgql

// Formatter provides a structured log formatter.
type Formatter struct{ Log logging.Logger }

// NewLogEntry creates a new log entry.
func (f *Formatter) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &entry{log: f.Log.WithValues(
		"id", middleware.GetReqID(r.Context()),
		"method", r.Method,
		"tls", r.TLS != nil,
		"host", r.Host,
		"uri", r.RequestURI,
		"protocol", r.Proto,
		"remote", r.RemoteAddr,
	)}
}

type entry struct{ log logging.Logger }

// Write writes a log entry.
func (e *entry) Write(status, bytes int, _ http.Header, elapsed time.Duration, _ any) {
	e.log.Debug("Handled request",
		"status", status,
		"bytes", bytes,
		"duration", elapsed,
	)
}

// Panic logs a panic.
func (e *entry) Panic(v any, stack []byte) {
	e.log.Debug("Paniced while handling request", "stack", stack, "panic", v)
}
