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
				log.Error().
					Err(startErr).
					Msgf("Error starting reader %v: %s", j, startErr)
				err = startErr
				return
			}
		}(i)
	}
	wg.Wait()
	log.Debug().Msgf("%v readers are closing", len(b.Readers))
	return err
}

func (b *Broadcaster) StartWriters(ctx context.Context) (err error) {
	for i := range b.Writers {
		go func(j int) {
			var upd model.Update
			feed, feedErr := b.Subscribe(ctx, fmt.Sprintf("writer %v %s", j, b.Writers[j].Name()))
			if feedErr != nil {
				err = feedErr
				return
			}
			for {
				select {
				case <-ctx.Done():
					return
				case upd = <-feed:
					chanName, found1 := b.FigiChannel[upd.Name]
					figiName, found2 := b.FigiName[upd.Name]
					if found1 && found2 {
						feedErr = b.Writers[j].Write(ctx, chanName, model.Update{
							Name:      figiName,
							Value:     upd.Value,
							Error:     upd.Error,
							Timestamp: upd.Timestamp,
						})
						if feedErr != nil {
							log.Error().Err(feedErr).
								Msgf("error publishing stock data to writer %v %s (%v) : %s",
									j, b.Writers[j].Name(), upd, feedErr,
								)
						}
					}
				}
			}
		}(i)
	}
	return err
}

func (b *Broadcaster) StartReaders(ctx context.Context) (err error) {
	if len(b.Writers) == 0 {
		return NoWritersError
	}
	if len(b.Readers) == 0 {
		return NoReadersError
	}
	if cap(b.Cord) < DefaultChannelBuffer {
		return fmt.Errorf("channel buffer is %v, while at least %v is recommended",
			cap(b.Cord), DefaultChannelBuffer)
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
