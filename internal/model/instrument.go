package model

type Instrument struct {
	FIGI           string
	Ticker         string
	InstrumentType string
	Name           string
	Lot            uint64
}
