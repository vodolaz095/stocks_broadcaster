package service

import (
	"github.com/vodolaz095/stocks_broadcaster/internal/model"
)

func (b *Broadcaster) Broadcast(upd model.Update) (subscribersNotified int) {
	var instrument model.Instrument
	for k := range b.subscribers {
		subscribersNotified += 1
		b.subscribers[k] <- upd
	}
	name, found := b.InstrumentGauges[upd.Name]
	if found {
		b.MetricsSet.GetOrCreateGauge(name, nil).Set(upd.Value)
	}
	name, found = b.InstrumentGauges["lot_"+upd.Name]
	if found {
		instrument, found = b.InstrumentCache[upd.Name]
		if found {
			b.MetricsSet.GetOrCreateGauge(name, nil).Set(upd.Value * float64(instrument.Lot))
		}
	}
	return subscribersNotified
}
