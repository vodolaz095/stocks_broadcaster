package redis

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/vodolaz095/stocks_broadcaster/model"
)

type Writer struct {
	Client *redis.Client
}

func (w *Writer) Name() string {
	return "redis publisher"
}

func (w *Writer) Ping(ctx context.Context) error {
	return w.Client.Ping(ctx).Err()
}

func (w *Writer) Close(_ context.Context) error {
	return w.Client.Close()
}

func (w *Writer) Write(ctx context.Context, channel string, upd model.Update) error {
	return w.Client.Publish(ctx, channel, upd.Pack()).Err()
}
