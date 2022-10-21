package api

import (
	"fmt"
	"net/http"
	"time"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/chi-middleware"
	oapifilter "github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"

	"github.com/upbound/upbound-go-api3/internal"
	apidemo "github.com/upbound/upbound-go-api3/internal/api/demo"
	// "github.com/upbound/upbound-go-api3/internal/client/auth"
	"github.com/upbound/upbound-go-api3/internal/log"
	srvdemo "github.com/upbound/upbound-go-api3/internal/server/api/demo"
	"github.com/upbound/upbound-go-api3/internal/server/metrics/otel"
	// "github.com/upbound/upbound-go-api3/internal/server/middleware"
)

// Server serves the Entities API.
func Server(opts internal.ServiceOptions) (*http.Server, error) {
	r := chi.NewRouter()
	r.Use(chimid.RequestLogger(&log.Formatter{Log: opts.Log}))
	r.Use(chimid.RedirectSlashes)
	r.Use(otel.Middleware)
	r.Use(chimid.Compress(5))
	// TODO(hasheddan): make this configurable and consider limiting
	// connections, not just requests.
	r.Use(chimid.Throttle(400))

	// For demo
	// // The auth manager is responsible for all authentication and authorization
	// // activity.
	// a := auth.New(opts.AuthHost, opts.PrivateHost, auth.WithLogger(opts.Log))

	// Add Demo API server to router.

	// Validate demo requests against OpenAPIv3 spec.
	repoSwagger, err := apidemo.GetSwagger()
	if err != nil {
		return nil, err
	}
	repoSwagger.Servers = nil

	// Override demo authentication because validator handles incorrectly.
	repoValidOpts := &oapimiddleware.Options{}
	repoValidOpts.Options.AuthenticationFunc = oapifilter.NoopAuthenticationFunc

	// Add demo API server to router.
	r.Group(func(r chi.Router) {
		r.Use(oapimiddleware.OapiRequestValidatorWithOptions(repoSwagger, repoValidOpts))

		// Remove for demo
		// // Authentication is required on all routes.
		// r.Use(middleware.NewAuthN(a, middleware.AuthNWithLogger(opts.Log)).Required)

		handlers := srvdemo.New(srvdemo.WithLogger(opts.Log))
		apidemo.HandlerFromMux(handlers, r)
	})

	return &http.Server{
		Handler:           r,
		Addr:              fmt.Sprintf(":%d", opts.APIPort),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}, nil
}
