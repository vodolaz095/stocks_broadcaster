package invest_api

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/go-investAPI/investapi"
	"github.com/vodolaz095/stocks_broadcaster/internal/model"
)

func (r *Reader) LoadInstruments(ctx context.Context, ids []string) ([]model.Instrument, error) {
	yes := true
	var err error
	var resp *investapi.FindInstrumentResponse
	var ret []model.Instrument
	for i := range ids {
		resp, err = r.Connection.InstrumentsServiceClient.FindInstrument(ctx, &investapi.FindInstrumentRequest{
			Query:                 ids[i],
			ApiTradeAvailableFlag: &yes,
		})
		if err != nil {
			log.Warn().Err(err).Msgf("Ошибка при поиске инструмента %s: %s", ids[i], err)
			continue
		}
		for j := range resp.Instruments {
			instrument := resp.GetInstruments()[j]
			log.Debug().Msgf("Instrument %s %s %s %s is found",
				instrument.GetFigi(),
				instrument.GetTicker(),
				instrument.GetInstrumentType(),
				instrument.GetName(),
			)
			ret = append(ret, model.Instrument{
				FIGI:           instrument.GetFigi(),
				Ticker:         instrument.GetTicker(),
				InstrumentType: instrument.GetInstrumentType(),
				Name:           instrument.GetName(),
				Lot:            uint64(instrument.GetLot()),
			})
		}
	}
	return ret, nil
}
