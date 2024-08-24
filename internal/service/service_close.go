package service

import (
	"context"

	"github.com/rs/zerolog/log"
)

func (b *Broadcaster) Close(ctx context.Context) (err error) {
	for i := range b.Readers {
		err = b.Readers[i].Close(ctx)
		if err != nil {
			log.Error().Err(err).Msgf("close: error closing reader %s: %s",
				b.Readers[i].Name(), err,
			)
			return err
		}
		log.Debug().Msgf("close: reader %s is terminated", b.Readers[i].Name())
	}
	for i := range b.Writers {
		err = b.Writers[i].Close(ctx)
		if err != nil {
			log.Error().Err(err).Msgf("ping: error closing writer %s: %s",
				b.Writers[i].Name(), err,
			)
			return err
		}
		log.Debug().Msgf("close: writer %s is terminated", b.Writers[i].Name())
	}
	log.Info().Msgf("close: system is stopped")
	return nil
}
