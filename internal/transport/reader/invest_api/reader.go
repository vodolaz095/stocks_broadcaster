package invest_api

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/vodolaz095/go-investAPI/investapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/vodolaz095/stocks_broadcaster/model"
)

const DefaultReadInterval = 10 * time.Millisecond

type Reader struct {
	Description  string
	Connection   *investapi.Client
	ReadInterval time.Duration
	Token        string
	Instruments  []string
}

func (r *Reader) Name() string {
	return "InvestAPI reader " + r.Description
}

func (r *Reader) Ping(ctx context.Context) error {
	return r.Connection.Ping(ctx)
}

func (r *Reader) Close(_ context.Context) (err error) {
	err = r.Connection.Connection.Close()
	code, ok := status.FromError(err)
	if ok {
		if code.Code() == codes.Canceled {
			return nil
		}
	}
	return err
}

func (r *Reader) Start(ctx context.Context, updateFeed chan model.Update) (err error) {
	var upd model.Update
	var instruments []*investapi.LastPriceInstrument
	for i := range r.Instruments {
		instruments = append(instruments, &investapi.LastPriceInstrument{Figi: r.Instruments[i]})
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
	stream, err := feed.MarketDataServerSideStream(ctx, &request)
	if err != nil {
		return fmt.Errorf("error subscribing to feed: %w", err)
	}
	var msg *investapi.MarketDataResponse
	var lastPrice *investapi.LastPrice
	go func() {
		<-ctx.Done()
		log.Info().Msgf("Closing grpc subscription for %v instruments", len(r.Instruments))
		// https://github.com/grpc/grpc-go/issues/3230#issuecomment-562061037
		r.Connection.Connection.Close()
	}()
	for {
		msg, err = stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Debug().Msgf("Closing grpc subscription loop")
				break
			}
			code, ok := status.FromError(err)
			if !ok {
				log.Error().Err(err).Msgf("subscription error: %s", err)
				break
			}
			if code.Code() == codes.Canceled {
				log.Debug().Msgf("Connection is canceled")
				return nil
			} else {
				log.Error().Err(err).Msgf("subscription error: %s", err)
			}
			break
		}
		log.Trace().Msgf("Message received: %s", msg.String())
		lastPrice = msg.GetLastPrice()
		if lastPrice != nil { // this is actual last price message
			log.Debug().Msgf("Instrument %s has last lot price %.4f",
				lastPrice.GetFigi(), lastPrice.GetPrice().ToFloat64())
			upd = model.Update{
				Name:      lastPrice.GetFigi(),
				Value:     lastPrice.Price.ToFloat64(),
				Error:     "",
				Timestamp: lastPrice.Time.AsTime(),
			}
			updateFeed <- upd
		}
	}
	return err
}
