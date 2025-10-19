package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"runtime"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/go-investAPI/investapi"
	"github.com/vodolaz095/pkg/stopper"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/vodolaz095/pkg/healthcheck"
	"github.com/vodolaz095/pkg/zerologger"
	"github.com/vodolaz095/stocks_broadcaster/config"
	"github.com/vodolaz095/stocks_broadcaster/internal/service"
	webserver "github.com/vodolaz095/stocks_broadcaster/internal/transport/http"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/reader"
	investapi_reader "github.com/vodolaz095/stocks_broadcaster/internal/transport/reader/invest_api"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/victoria_metrics"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/writer"
	redisWriter "github.com/vodolaz095/stocks_broadcaster/internal/transport/writer/redis"
	"github.com/vodolaz095/stocks_broadcaster/model"
)

var Version = "development"

func main() {
	var err error
	mainCtx, cancel := stopper.New()
	defer cancel()
	flag.Parse()

	// load config
	if len(flag.Args()) != 1 {
		log.Fatal().Msgf("please, provide path to config as 1st argument")
	}
	pathToConfig := flag.Args()[0]
	cfg, err := config.LoadFromFile(pathToConfig)
	if err != nil {
		log.Fatal().Err(err).
			Msgf("error loading config from %s: %s", pathToConfig, err)
	}
	validate := validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(cfg)
	if err != nil {
		log.Fatal().Err(err).
			Msgf("error validating configuration file %s: %s", pathToConfig, err)
	}

	// set logging
	zerologger.Configure(cfg.Log)

	log.Info().Msgf("Starting StockBroadcaster version %s. GOOS: %s. ARCH: %s. Go Version: %s. Please, report bugs here: %s",
		Version, runtime.GOOS, runtime.GOARCH, runtime.Version(),
		"https://github.com/vodolaz095/stocks_broadcaster/issues",
	)

	metricsSet := metrics.NewSet()

	// configure readers
	var readers []reader.StocksReader
	for i := range cfg.Inputs {
		var dialer *net.Dialer
		if cfg.Inputs[i].LocalAddr != "" {
			// make connections only from one of local network addresses
			dialer = &net.Dialer{
				LocalAddr: &net.TCPAddr{
					IP: net.ParseIP(cfg.Inputs[i].LocalAddr),
				},
			}
			log.Info().Msgf("Reader %v %s uses local address %s to dial invest API",
				i, cfg.Inputs[i].Name, cfg.Inputs[i].LocalAddr)
		} else {
			// make kernel choose local network interface to dial
			dialer = &net.Dialer{}
		}
		investApiClient, err1 := investapi.NewWithOpts(
			cfg.Inputs[i].Token,
			investapi.DefaultEndpoint,
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, "tcp", addr)
			}),
		)
		if err != nil {
			log.Fatal().
				Err(err1).
				Msgf("error connecting invest api: %s", err1)
		}
		readers = append(readers, &investapi_reader.Reader{
			MetricsSet:   metricsSet,
			Description:  cfg.Inputs[i].Name,
			Connection:   investApiClient,
			ReadInterval: investapi_reader.DefaultReadInterval,
			Instruments:  cfg.Inputs[i].Figis,
		})
		log.Info().Msgf("Reader %v-%s for %v instruments is prepared", i, cfg.Inputs[i].Name, len(cfg.Inputs[i].Figis))
	}

	// configure writers
	var writers []writer.StocksWriter
	for i := range cfg.Outputs {
		redisOpts, err2 := redis.ParseURL(cfg.Outputs[i].RedisURL)
		if err2 != nil {
			log.Fatal().Err(err2).Msgf("error parsing redis connection string %s from %v: %s",
				cfg.Outputs[i].RedisURL, i, err2)
		}
		writers = append(writers, &redisWriter.Writer{
			Description: cfg.Outputs[i].Name,
			Client:      redis.NewClient(redisOpts),
		})
		log.Info().Msgf("Writer %v-%s is prepared", i, cfg.Outputs[i].Name)
	}

	// configure service
	srv := service.Broadcaster{
		MetricsSet:       metricsSet,
		FigiName:         make(map[string]string, 0),
		FigiChannel:      make(map[string]string, 0),
		Cord:             make(chan model.Update, service.DefaultChannelBuffer),
		Readers:          readers,
		Writers:          writers,
		InstrumentGauges: make(map[string]string, 0),
	}
	// configure service routing
	for i := range cfg.Instruments {
		srv.FigiName[cfg.Instruments[i].FIGI] = cfg.Instruments[i].Name
		srv.FigiChannel[cfg.Instruments[i].FIGI] = cfg.Instruments[i].Channel
		srv.InstrumentGauges[cfg.Instruments[i].FIGI] = fmt.Sprintf("%s{figi=%q}",
			cfg.Instruments[i].Name, cfg.Instruments[i].FIGI)
	}
	// set systemd watchdog
	systemdWatchdogEnabled, err := healthcheck.Ready()
	if err != nil {
		log.Error().Err(err).
			Msgf("%s: while notifying systemd on application ready", err)
	}
	log.Info().Msgf("Starting service with %v readers and %v writers", len(srv.Readers), len(srv.Writers))
	eg, ctx := errgroup.WithContext(mainCtx)
	eg.Go(func() error {
		if !systemdWatchdogEnabled {
			log.Warn().Msgf("Systemd watchdog disabled - application can work unstable in systemd environment")
			return nil
		}
		log.Debug().Msgf("Watchdog enabled")
		return healthcheck.StartWatchDog(ctx, []healthcheck.Pinger{&srv})
	})
	eg.Go(func() error {
		return srv.StartWriters(ctx)
	})
	eg.Go(func() error {
		return srv.StartReaders(ctx)
	})
	// change systemd status
	if systemdWatchdogEnabled {
		// https://www.freedesktop.org/software/systemd/man/latest/sd_notify.html#STATUS=%E2%80%A6
		err = healthcheck.SetStatus("Broadcasting stock data...")
		if err != nil {
			log.Warn().Err(err).Msgf("Error setting systemd unit status")
		}
	}
	eg.Go(func() error {
		if !cfg.Webserver.Enabled {
			log.Debug().Msgf("Webserver is not enabled")
			return nil
		}
		ws := webserver.WebServer{
			Service:              &srv,
			ExposeRuntimeMetrics: cfg.Webserver.ExposeRuntimeMetrics,
			Network:              cfg.Webserver.Network,
			Listen:               cfg.Webserver.Listen,
			Socket:               cfg.Webserver.Socket,
		}
		return ws.Start(ctx)
	})

	// push metrics to victoria metrics
	for i := range cfg.VictoriaMetricsDatabases {
		vm := victoria_metrics.Writer{
			Endpoint: cfg.VictoriaMetricsDatabases[i].Endpoint,
			Headers:  cfg.VictoriaMetricsDatabases[i].Headers,
			Labels:   cfg.VictoriaMetricsDatabases[i].Labels,
			Interval: cfg.VictoriaMetricsDatabases[i].Interval,
		}
		eg.Go(func() error {
			return vm.Start(ctx, srv.MetricsSet)
		})
		if cfg.VictoriaMetricsDatabases[i].ExposeRuntimeMetrics {
			eg.Go(func() error {
				return vm.StartSendingRuntimeMetrics(ctx)
			})
		}
	}

	// main loop
	err = eg.Wait()
	if err != nil {
		log.Error().Err(err).Msgf("Error starting system: %s", err)
	}

	// termination
	terminationContext, terminationContextCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer terminationContextCancel()
	err = srv.Close(terminationContext)
	if err != nil {
		log.Error().Err(err).
			Msgf("Error terminating application, something can be broken: %s", err)
	} else {
		log.Info().Msgf("Stocks Broadcaster is terminated.")
	}
}
