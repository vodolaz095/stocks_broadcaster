package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/stocks_broadcaster/internal/service"
)

type WebServer struct {
	Service  *service.Broadcaster
	Listener net.Listener
	Network  string
	Listen   string
	Socket   string
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
	})
	srv := http.Server{
		Handler:      mux,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}
	return srv.Serve(listener)
}
