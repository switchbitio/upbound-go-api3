package otel

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/unit"
)

const (
	Success   = "success"
	Malformed = "malformed"
	Rejected  = "rejected"
)

var (
	meter = global.GetMeterProvider().Meter("upbound-go-api3")

	reqStarted = metric.Must(meter).NewInt64Counter("http.request.started.total",
		metric.WithDescription("Total number of http requests started."),
		metric.WithUnit(unit.Dimensionless))

	reqCompleted = metric.Must(meter).NewInt64Counter("http.request.completed.total",
		metric.WithDescription("Total number of http requests completed."),
		metric.WithUnit(unit.Dimensionless))

	reqDuration = metric.Must(meter).NewFloat64Histogram("http.request.duration.ms",
		metric.WithDescription("Time between receiving and responding to an http request."),
		metric.WithUnit(unit.Milliseconds))

	productMetricSubmitted = metric.Must(meter).NewInt64Counter("prodmetric.submitted",
		metric.WithDescription("Total number of product metrics submitted."),
		metric.WithUnit(unit.Dimensionless))
)

// Middleware records metrics for HTTP handlers.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqStarted.Add(context.Background(), 1, HTTPServerMetricAttributesFromHTTPRequest(r)...)
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			reqCompleted.Add(context.Background(), 1, HTTPServerMetricAttributesFromHTTPResponse(r, ww)...)
			reqDuration.Record(context.Background(), float64(time.Since(t1).Milliseconds()), HTTPServerMetricAttributesFromHTTPResponse(r, ww)...)
		}()
		next.ServeHTTP(ww, r)
	})
}

// HTTPServerMetricAttributesFromHTTPRequest constructs default attributes for
// an HTTP request.
func HTTPServerMetricAttributesFromHTTPRequest(r *http.Request) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("http.method", r.Method),
		attribute.String("http.host", r.Host),
		attribute.Bool("http.tls", r.TLS != nil),
	}
}

// HTTPServerMetricAttributesFromHTTPResponse constructs default attributes for
// an HTTP response.
func HTTPServerMetricAttributesFromHTTPResponse(r *http.Request, w middleware.WrapResponseWriter) []attribute.KeyValue {
	return append(HTTPServerMetricAttributesFromHTTPRequest(r), attribute.Int("http.status_code", w.Status()))
}

// ProductMetricSubmit records an product metric submission.
func ProductMetricSubmit(account, repository string, success bool) {
	productMetricSubmitted.Add(context.Background(), 1, []attribute.KeyValue{
		attribute.String("prodmetric.account", account),
		attribute.String("prodmetric.repository", repository),
		attribute.Bool("prodmetric.success", success),
	}...)
}
