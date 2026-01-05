package reader

import (
	"context"

	"github.com/vodolaz095/stocks_broadcaster/internal/model"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport"
)

type StocksReader interface {
	transport.Transport
	LoadInstruments(context.Context, []string) ([]model.Instrument, error)
	Start(context.Context, chan model.Update) error
}
