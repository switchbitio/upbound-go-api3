package entities

import (
	"net/http"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

// GetV1Demo - [/v1/demo] - Demo
func (h *Demo) GetV1Demo(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("Demo"))
}

// Demo implements the demo OpenAPI spec.
type Demo struct {
	log logging.Logger
}

// Opt sets an option on the Demo API.
type Opt func(p *Demo)

// WithLogger sets the logger for the Demo API.
func WithLogger(l logging.Logger) Opt {
	return func(r *Demo) {
		r.log = l
	}
}

func New(opts ...Opt) *Demo {
	r := &Demo{
		log: logging.NewNopLogger(),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}
