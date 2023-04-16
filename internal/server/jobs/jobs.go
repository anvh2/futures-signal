package jobs

import (
	"github.com/anvh2/futures-signal/internal/cache"
	"github.com/anvh2/futures-signal/internal/logger"
	"github.com/anvh2/futures-signal/internal/services/binance"
	"github.com/anvh2/futures-signal/internal/services/telegram"
	"github.com/anvh2/futures-signal/internal/worker"
)

var (
	blacklist = map[string]bool{}
)

type Jobs struct {
	logger        *logger.Logger
	binance       *binance.Binance
	notify        *telegram.TelegramBot
	worker        *worker.Worker
	cache         cache.Basic
	marketCache   cache.Market
	exchangeCache cache.Exchange
	retryChannel  chan *retry
	quitChannel   chan struct{}
}

func New(
	logger *logger.Logger,
	binance *binance.Binance,
	notify *telegram.TelegramBot,
	worker *worker.Worker,
	cache cache.Basic,
	marketCache cache.Market,
	exchangeCache cache.Exchange,
	quitChannel chan struct{},
) *Jobs {
	return &Jobs{
		logger:        logger,
		binance:       binance,
		notify:        notify,
		worker:        worker,
		cache:         cache,
		marketCache:   marketCache,
		exchangeCache: exchangeCache,
		retryChannel:  make(chan *retry, 1000),
		quitChannel:   quitChannel,
	}
}
