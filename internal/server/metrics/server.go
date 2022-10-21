package metrics

import (
	"fmt"
	"net/http"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/go-chi/chi/v5"
	chimid "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	"go.opentelemetry.io/otel/sdk/metric/export/aggregation"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"

	"github.com/upbound/upbound-go-api3/internal"
	"github.com/upbound/upbound-go-api3/internal/log"
)

// Server serves the metrics API.
func Server(opts internal.MetricsOptions, logger logging.Logger) (*http.Server, error) {
	config := prometheus.Config{}
	ctrl := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
		controller.WithResource(resource.NewSchemaless(attribute.String("upbound-go-api3.name", "upbound-go-api3"))),
	)
	exporter, err := prometheus.New(config, ctrl)
	if err != nil {
		return nil, err
	}

	// Set prometheus exporter as global meter provider to allow access from
	// other packages.
	global.SetMeterProvider(exporter.MeterProvider())

	mr := chi.NewRouter()
	mr.Use(chimid.RequestLogger(&log.Formatter{Log: logger}))
	mr.Handle("/metrics", exporter)
	return &http.Server{
		Handler:           mr,
		Addr:              fmt.Sprintf(":%d", opts.MetricsPort),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
	}, nil
}
