package service

import (
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/reader"
	"github.com/vodolaz095/stocks_broadcaster/internal/transport/writer"
	"github.com/vodolaz095/stocks_broadcaster/model"
)

const DefaultChannelBuffer = 100

type Broadcaster struct {
	Cord    chan model.Update
	Readers []reader.StocksReader
	Writers []writer.StocksWriter
}
