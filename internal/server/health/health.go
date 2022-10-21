package health

import (
	"net/http"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

// GetLiveness gets the servic liveness.
func (h *Probes) GetLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// GetReadiness gets the upbound-go-api3 readiness.
func (h *Probes) GetReadiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// Probes indicates the health of a upbound-go-api3.
type Probes struct {
	log logging.Logger
}

// Opt sets an option on the probes API.
type Opt func(p *Probes)

// WithLogger sets the logger for the probes API.
func WithLogger(l logging.Logger) Opt {
	return func(p *Probes) {
		p.log = l
	}
}

// New constructs a new probes API.
func New(opts ...Opt) *Probes {
	p := &Probes{
		log: logging.NewNopLogger(),
	}

	for _, o := range opts {
		o(p)
	}
	return p
}
