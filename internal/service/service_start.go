package service

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/stocks_broadcaster/model"
)

func (b *Broadcaster) startReaders(ctx context.Context) (err error) {
	if len(b.Readers) == 0 {
		return fmt.Errorf("no readers defined")
	}
	wg := sync.WaitGroup{}

	for i := range b.Readers {
		log.Debug().Msgf("Preparing to start reader %v %s...", i, b.Readers[i].Name())
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			startErr := b.Readers[j].Start(ctx, b.Cord)
			if startErr != nil {
				log.Error().Err(startErr).Msgf("Error starting reader %v: %s", j, startErr)
				err = startErr
				return
			}
			log.Debug().Msgf("Reader #%v %s is started!", j, b.Readers[j].Name())
		}(i)
	}
	wg.Wait()
	return err
}

func (b *Broadcaster) Start(ctx context.Context) (err error) {
	if len(b.Writers) == 0 {
		return NoWritersError
	}
	if len(b.Readers) == 0 {
		return NoReadersError
	}
	if cap(b.Cord) < DefaultChannelBuffer {
		return fmt.Errorf("channel buffer is %v, while at least 100 is recommended", cap(b.Cord))
	}
	go func() {
		var upd model.Update
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msgf("Closing broadcasting...")
				return
			case upd = <-b.Cord:
				b.Broadcast(upd)
			}
		}
	}()
	err = b.startReaders(ctx)
	return err
}
