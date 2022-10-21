package runtime

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"golang.org/x/sync/errgroup"
)

const (
	errFailedServing  = "failed serving"
	errFailedShutdown = "failed shutdown"
)

type srv struct {
	name string
	log  logging.Logger
}

// SrvOpt modifies a server.
type SrvOpt func(*srv)

// WithName sets the server name.
func WithName(n string) SrvOpt {
	return func(s *srv) {
		s.name = n
	}
}

// WithLogger sets the server logger.
func WithLogger(l logging.Logger) SrvOpt {
	return func(s *srv) {
		s.log = l
	}
}

// StartServer - run an http server with shutdown monitoring
func StartServer(h *http.Server, g *errgroup.Group, sd <-chan struct{}, opts ...SrvOpt) {
	s := &srv{
		name: "default",
		log:  logging.NewNopLogger(),
	}
	for _, o := range opts {
		o(s)
	}
	log := s.log.WithValues("Server", s.name, "Address", h.Addr)
	g.Go(func() error {
		log.Debug("Starting server.")
		if err := h.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Info(errFailedServing, "Error", err)
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-sd
		to, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		log.Debug("Shutting down server.")
		if err := h.Shutdown(to); err != nil {
			log.Info(errFailedShutdown, "Error", err)
			return err
		}
		log.Debug("Shutdown successful.")
		return nil
	})
}

// StartUpDown - run a generic blocking function in thread with shutdown monitoring
func StartUpDown(up func() error, down func(context.Context) error, g *errgroup.Group, sd <-chan struct{}, opts ...SrvOpt) {
	s := &srv{
		name: "default",
		log:  logging.NewNopLogger(),
	}
	for _, o := range opts {
		o(s)
	}
	log := s.log.WithValues("Server", s.name)
	g.Go(func() error {
		log.Debug("Starting server.")
		if err := up(); err != nil {
			log.Info(errFailedServing, "Error", err)
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-sd
		to, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		log.Debug("Shutting down server.")
		if down != nil {
			if err := down(to); err != nil {
				log.Info(errFailedShutdown, "Error", err)
				return err
			}
		}
		log.Debug("Shutdown successful.")
		return nil
	})
}
