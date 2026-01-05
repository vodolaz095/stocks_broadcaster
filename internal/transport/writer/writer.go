package writer

import (
	"context"

	"github.com/vodolaz095/stocks_broadcaster/internal/model"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport"
)

type StocksWriter interface {
	transport.Transport
	Write(ctx context.Context, channel string, upd model.Update) error
}
