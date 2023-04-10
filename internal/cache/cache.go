package cache

import (
	"github.com/anvh2/futures-signal/internal/cache/exchange"
	"github.com/anvh2/futures-signal/internal/cache/market"
)

//go:generate moq -pkg cachemock -out ./mocks/market_mock.go . Market
type Market interface {
	CandleSummary(symbol string) (*market.CandleSummary, error)
	CreateSummary(symbol string) *market.CandleSummary
	UpdateSummary(symbol string) *market.CandleSummary
}

//go:generate moq -pkg cachemock -out ./mocks/exchange_mock.go . Exchange
type Exchange interface {
	Set(symbols []*exchange.Symbol)
	Get(symbol string) (*exchange.Symbol, error)
	Symbols() []string
}

type Basic interface {
	Set(key string, value interface{})
	Get(key string) interface{}
	SetEX(key string, value interface{}) (interface{}, bool)
}
