package internal

import (
	"net/url"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
)

// CommonOptions are options exposed by all services.
type CommonOptions struct {
	Log          logging.Logger `kong:"-"`
	Debug        bool           `name:"debug" env:"DEBUG" short:"d" default:"false" help:"Run with debug logging."`
	DevMode      bool           `name:"dev-mode" env:"DEV_MODE" default:"false" help:"Enables logging dev mode."`
	EnableGZip   bool           `name:"enable-gzip" env:"ENABLE_GZIP" default:"true" help:"Enable gzip compression. Default value = true"`
	PrivatePort  int            `default:"8089" help:"Port for private API server."`
	IsEnterprise bool           `kong:"-"`
	MetricsOptions
}

// ProductMetricsOptions are common options for consumers of the accounts upbound-go-api3.
type ProductMetricsOptions struct {
	Host url.URL `name:"product-metrics-host" default:"http://product-metrics-private:8080" help:"Product Metrics upbound-go-api3 host."`
}

// MetricsOptions options related to prometheus metrics server
type MetricsOptions struct {
	MetricsPort int  `default:"8085" help:"Port for metrics server."`
	Metrics     bool `name:"metrics" default:"true" negatable:"" help:"Enable Prometheus metrics exporter."`
}
