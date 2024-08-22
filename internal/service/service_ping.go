package service

import (
	"context"

	"github.com/rs/zerolog/log"
)

func (b *Broadcaster) Ping(ctx context.Context) (err error) {
	for i := range b.Readers {
		err = b.Readers[i].Ping(ctx)
		if err != nil {
			log.Error().Err(err).Msgf("ping: error pinging reader %s: %s",
				b.Readers[i].Name(), err,
			)
			return err
		}
		log.Debug().Msgf("ping: reader %s is online", b.Readers[i].Name())
	}
	for i := range b.Writers {
		err = b.Writers[i].Ping(ctx)
		if err != nil {
			log.Error().Err(err).Msgf("ping: error pinging writer %s: %s",
				b.Writers[i].Name(), err,
			)
			return err
		}
		log.Debug().Msgf("ping: writer %s is online", b.Writers[i].Name())
	}
	log.Debug().Msgf("ping: system online")
	return nil
}
