package internal

import "net/url"

// ServiceOptions defines the available set of configuration options available
// for the API.
type ServiceOptions struct {
	APIPort int  `default:"8081" help:"Port for API server."`
	API     bool `name:"api" default:"true" negatable:"" help:"Run with the API server enabled."`

	AuthHost    url.URL `default:"http://api-private-auth:8081" help:"Auth upbound-go-api3 host."`
	PrivateHost url.URL `default:"http://api-private:8081" help:"Private upbound-go-api3 host."`

	CommonOptions

	ProductMetricsOptions
}
