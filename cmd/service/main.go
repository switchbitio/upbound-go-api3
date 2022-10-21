package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/alecthomas/kong"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/upbound/upbound-go-api3/internal"
	"github.com/upbound/upbound-go-api3/internal/runtime"
	"github.com/upbound/upbound-go-api3/internal/server/api"
	"github.com/upbound/upbound-go-api3/internal/server/metrics"
	"github.com/upbound/upbound-go-api3/internal/server/private"
)

func main() {
	opts := internal.ServiceOptions{}

	ctx := kong.Parse(&opts, kong.Name("upbound-go-api3"),
		kong.Description("Upbound upbound-go-api3"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:   true,
			FlagsLast: true,
			Summary:   true,
		}))

	// specify logging options
	zapOpts := []zap.Opts{}
	if opts.Debug {
		zapOpts = append(zapOpts, zap.Level(zapcore.DebugLevel))
	}
	if opts.DevMode {
		zapOpts = append(zapOpts, zap.UseDevMode(true))
	}

	zl := zap.New(zapOpts...)
	opts.Log = logging.NewLogrLogger(zl.WithName("upbound-go-api3"))

	if err := run(opts); err != nil {
		ctx.FatalIfErrorf(err)
	}
	ctx.Exit(0)
}

// Run Instantiates and runs the services
func run(opts internal.ServiceOptions) error { // nolint: gocyclo
	g, ctx := errgroup.WithContext(context.Background())
	done := make(chan struct{})

	if opts.API {
		apiServer, err := api.Server(opts)
		if err != nil {
			return err
		}
		runtime.StartServer(apiServer, g, done, runtime.WithLogger(opts.Log), runtime.WithName("api"))
	}

	if opts.Metrics {
		metricsServer, err := metrics.Server(opts.MetricsOptions, opts.Log)
		if err != nil {
			return err
		}
		runtime.StartServer(metricsServer, g, done, runtime.WithLogger(opts.Log), runtime.WithName("metrics"))
	}

	privateServer, err := private.Server(opts)
	if err != nil {
		return err
	}
	runtime.StartServer(privateServer, g, done, runtime.WithLogger(opts.Log), runtime.WithName("private"))

	sigint := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM)

		// Block until signal or one server errors.
		select {
		case <-sigint:
		case <-ctx.Done():
		}

		// Signal all servers to shutdown.
		close(done)
	}()
	return g.Wait()
}
