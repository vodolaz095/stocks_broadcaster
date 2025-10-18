package service

import (
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/reader"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/writer"
	"github.com/vodolaz095/stocks_broadcaster/model"

	"github.com/VictoriaMetrics/metrics"
)

const DefaultChannelBuffer = 100

type Broadcaster struct {
	FigiName         map[string]string
	FigiChannel      map[string]string
	InstrumentGauges map[string]string

	Cord       chan model.Update
	Readers    []reader.StocksReader
	Writers    []writer.StocksWriter
	MetricsSet *metrics.Set

	// subscribers are used to send updates to different transports - redis publishers,
	// influx, etc...
	subscribers map[string]chan model.Update
}
