package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/go-investAPI/investapi"

	"github.com/vodolaz095/stocks_broadcaster/config"
	"github.com/vodolaz095/stocks_broadcaster/internal/service"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/reader"
	investapi_reader "github.com/vodolaz095/stocks_broadcaster/internal/transport/reader/invest_api"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/writer"
	redisWriter "github.com/vodolaz095/stocks_broadcaster/internal/transport/writer/redis"
	"github.com/vodolaz095/stocks_broadcaster/model"
	"github.com/vodolaz095/stocks_broadcaster/pkg/healthcheck"
	"github.com/vodolaz095/stocks_broadcaster/pkg/zerologger"
)

var Version = "development"

func main() {
	var err error
	wg := sync.WaitGroup{}
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
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
		Version,
		runtime.GOOS, runtime.GOARCH, runtime.Version(),
		"https://github.com/vodolaz095/stocks_broadcaster/issues",
	)

	// configure readers
	investApiClient, err := investapi.New(cfg.Token)
	if err != nil {
		log.Fatal().Err(err).Msgf("error connecting invest api: %s", err)
	}
	iaReader := investapi_reader.Reader{
		Connection:   investApiClient,
		ReadInterval: investapi_reader.DefaultReadInterval,
		Instruments:  cfg.Instruments,
	}

	// configure writers
	redisOpts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msgf("error parsing redis connection string %s: %s", cfg.RedisURL, err)
	}
	client := redis.NewClient(redisOpts)
	rw := redisWriter.Writer{
		Client: client,
	}

	// configure service
	srv := service.Broadcaster{
		FigiName:    make(map[string]string, 0),
		FigiChannel: make(map[string]string, 0),
		Cord:        make(chan model.Update, service.DefaultChannelBuffer),
		Readers:     []reader.StocksReader{&iaReader}, // todo - MORE!
		Writers:     []writer.StocksWriter{&rw},       // todo - MORE!
	}
	// configure service routing
	for i := range cfg.Instruments {
		srv.FigiName[cfg.Instruments[i].FIGI] = cfg.Instruments[i].Name
		srv.FigiChannel[cfg.Instruments[i].FIGI] = cfg.Instruments[i].Channel
	}

	// set systemd watchdog
	systemdWatchdogEnabled, err := healthcheck.Ready()
	if err != nil {
		log.Error().Err(err).
			Msgf("%s: while notifying systemd on application ready", err)
	}
	if systemdWatchdogEnabled {
		go func() {
			log.Debug().Msgf("Watchdog enabled")
			errWD := healthcheck.StartWatchDog(ctx, []healthcheck.Pinger{
				&srv,
			})
			if errWD != nil {
				log.Error().
					Err(err).
					Msgf("%s : while starting watchdog", err)
			}
		}()
	} else {
		log.Warn().Msgf("Systemd watchdog disabled - application can work unstable in systemd environment")
	}

	// change systemd status
	if systemdWatchdogEnabled {
		// https://www.freedesktop.org/software/systemd/man/latest/sd_notify.html#STATUS=%E2%80%A6
		err = healthcheck.SetStatus("Broadcasting stock data...")
		if err != nil {
			log.Warn().Err(err).Msgf("Error setting systemd unit status")
		}
	}

	// handle signals
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGABRT,
	)
	go func() {
		s := <-sigc
		log.Info().Msgf("Signal %s is received", s.String())
		wg.Done()
		cancel()
	}()

	// main loop
	err = srv.Start(ctx)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error starting broadcaster: %s", err)
	}

	// closing
	wg.Wait()
	terminationContext, terminationContextCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer terminationContextCancel()
	err = srv.Close(terminationContext)
	if err != nil {
		log.Error().Err(err).
			Msgf("Error terminating application, something can be broken: %s", err)
	}
	log.Info().Msgf("Stocks Broadcaster is terminated.")
}
