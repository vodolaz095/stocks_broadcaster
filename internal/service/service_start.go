package service

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

func (b *Broadcaster) Start(ctx context.Context) (err error) {
	//if len(b.Writers) == 0 {
	//	return fmt.Errorf("no writers defined")
	//}
	if len(b.Readers) == 0 {
		return fmt.Errorf("no readers defined")
	}
	if cap(b.Cord) < DefaultChannelBuffer {
		return fmt.Errorf("channel buffer is %v, while at least 100 is recommended", cap(b.Cord))
	}
	for i := range b.Writers {
		log.Debug().Msgf("Preparing to start writer %v %s...", i, b.Writers[i].Name())
		go b.Writers[i].Start(ctx, b.Cord)
	}
	for i := range b.Readers {
		log.Debug().Msgf("Preparing to start reader %v %s...", i, b.Readers[i].Name())
		go b.Readers[i].Start(ctx, b.Cord)
	}
	return nil
}
