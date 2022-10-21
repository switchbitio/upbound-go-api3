package private

import (
	"fmt"
	"net/http"
	"time"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	oapifilter "github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"

	"github.com/upbound/upbound-go-api3/internal"
	healthapi "github.com/upbound/upbound-go-api3/internal/api/health"
	"github.com/upbound/upbound-go-api3/internal/server/health"
)

// Server is a private API server.
func Server(opts internal.ServiceOptions) (*http.Server, error) {
	r := chi.NewRouter()
	r.Use(chimid.RedirectSlashes)
	r.Use(chimid.Compress(5))

	// Validate health requests against OpenAPIv3 spec.
	healthSwagger, err := healthapi.GetSwagger()
	if err != nil {
		return nil, err
	}
	healthSwagger.Servers = nil

	// Override authentication because validator handles incorrectly.
	healthValidOpts := &oapimiddleware.Options{}
	healthValidOpts.Options.AuthenticationFunc = oapifilter.NoopAuthenticationFunc

	r.Group(func(r chi.Router) {
		r.Use(oapimiddleware.OapiRequestValidatorWithOptions(healthSwagger, healthValidOpts))
		handlers := health.New(health.WithLogger(opts.Log))
		healthapi.HandlerFromMux(handlers, r)
	})

	return &http.Server{
		Handler:           r,
		Addr:              fmt.Sprintf(":%d", opts.PrivatePort),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}, nil
}
