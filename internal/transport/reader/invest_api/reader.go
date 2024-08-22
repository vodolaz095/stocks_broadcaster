package invest_api

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/go-investAPI/investapi"
	"github.com/vodolaz095/stocks_broadcaster/config"
	"github.com/vodolaz095/stocks_broadcaster/model"
)

const DefaultReadInterval = 10 * time.Millisecond

type Reader struct {
	Connection   *investapi.Client
	ReadInterval time.Duration
	Token        string
	Instruments  []config.Instrument
}

func (r *Reader) Name() string {
	return "InvestAPI reader"
}

func (r *Reader) Ping(ctx context.Context) error {
	return r.Connection.Ping(ctx)
}

func (r *Reader) Close(_ context.Context) error {
	return r.Connection.Connection.Close()
}

func (r *Reader) Start(ctx context.Context, updateFeed chan model.Update) (err error) {
	var instruments []*investapi.LastPriceInstrument
	for i := range r.Instruments {
		instruments = append(instruments, &investapi.LastPriceInstrument{Figi: r.Instruments[i].FIGI})
	}
	//  подписываемся на цену крайней сделки по акциям
	request := investapi.MarketDataServerSideStreamRequest{
		SubscribeCandlesRequest:   nil,
		SubscribeOrderBookRequest: nil,
		SubscribeTradesRequest:    nil,
		SubscribeInfoRequest:      nil,
		SubscribeLastPriceRequest: &investapi.SubscribeLastPriceRequest{
			SubscriptionAction: investapi.SubscriptionAction_SUBSCRIPTION_ACTION_SUBSCRIBE,
			Instruments:        instruments,
		},
	}
	feed := investapi.NewMarketDataStreamServiceClient(r.Connection.Connection)
	stream, err := feed.MarketDataServerSideStream(context.TODO(), &request)
	if err != nil {
		return fmt.Errorf("error subscribing to feed: %w", err)
	}
	defer stream.CloseSend()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	defer log.Info().Msgf("Closing subscription for %v instruments", len(r.Instruments))
	var msg *investapi.MarketDataResponse
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			msg, err = stream.Recv()
			if err != nil {
				log.Error().Err(err).Msgf("subscription error: %s", err)
				continue
			}
			log.Trace().Msgf("Message received: %s", msg.String())
		}
	}
}
