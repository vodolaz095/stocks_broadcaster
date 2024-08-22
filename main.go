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
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/stocks_broadcaster/internal/service"
	"github.com/vodolaz095/stocks_broadcaster/model"
	"github.com/vodolaz095/stocks_broadcaster/pkg/healthcheck"

	"github.com/vodolaz095/stocks_broadcaster/config"
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

	// configure writers

	// configure readers

	// configure service
	srv := service.Broadcaster{
		Cord:    make(chan model.Update, service.DefaultChannelBuffer),
		Readers: nil,
		Writers: nil,
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
