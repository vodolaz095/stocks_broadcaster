package service

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/vodolaz095/stocks_broadcaster/model"
)

func (b *Broadcaster) Subscribe(ctx context.Context, name string) (chan model.Update, error) {
	_, found := b.subscribers[name]
	if found {
		return nil, DuplicateSubscriberError
	}
	log.Debug().Msgf("Creating subscription channel for %s...", name)
	ch := make(chan model.Update, DefaultSubscriptionChannelChannelDepth)
	if b.subscribers == nil {
		b.subscribers = make(map[string]chan model.Update, 0)
	}
	b.subscribers[name] = ch
	go func() {
		<-ctx.Done()
		log.Debug().Msgf("Closing subscription channel for %s...", name)
		close(ch)
		delete(b.subscribers, name)
		log.Debug().Msgf("Subscription channel for %s is closed", name)
	}()

	return ch, nil
}
