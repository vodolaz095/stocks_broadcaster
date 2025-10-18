package service

import (
	"github.com/vodolaz095/stocks_broadcaster/model"
)

func (b *Broadcaster) Broadcast(upd model.Update) (subscribersNotified int) {
	for k := range b.subscribers {
		subscribersNotified += 1
		b.subscribers[k] <- upd
	}
	name, found := b.InstrumentGauges[upd.Name]
	if found {
		b.MetricsSet.GetOrCreateGauge(name, nil).Set(upd.Value)
	}
	return subscribersNotified
}
