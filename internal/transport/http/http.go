package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/stocks_broadcaster/internal/service"
)

type WebServer struct {
	Service              *service.Broadcaster
	Listener             net.Listener
	ExposeRuntimeMetrics bool
	Network              string
	Listen               string
	Socket               string
}

func (ws *WebServer) Name() string {
	return fmt.Sprintf("webserver on %s", ws.Listener.Addr())
}

func (ws *WebServer) Ping(context.Context) error {
	return nil
}

func (ws *WebServer) Close(context.Context) error {
	return ws.Listener.Close()
}

func (ws *WebServer) Start(ctx context.Context) error {
	var where string
	if ws.Network == "unix" {
		where = ws.Socket
	} else {
		where = ws.Listen
	}
	listener, err := net.Listen(ws.Network, where)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		log.Debug().Msgf("Closing HTTP server on %s %s...", ws.Network, where)
		listener.Close()
	}()
	log.Info().Msgf("Starting HTTP server on %s %s", ws.Network, where)

	mux := http.NewServeMux()
	mux.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("ready"))
		writer.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		errPing := ws.Service.Ping(request.Context())
		if errPing != nil {
			writer.Write([]byte(errPing.Error()))
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
		writer.Write([]byte("all systems online"))
		writer.WriteHeader(http.StatusOK)
		return
	})
	mux.HandleFunc("/metrics", func(writer http.ResponseWriter, _ *http.Request) {
		ws.Service.MetricsSet.WritePrometheus(writer)
		if ws.ExposeRuntimeMetrics {
			metrics.WritePrometheus(writer, true)
		}
		writer.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	srv := http.Server{
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	err = srv.Serve(listener)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return nil
		}
		return err
	}
	return err
}
