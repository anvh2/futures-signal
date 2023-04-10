package server

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anvh2/futures-signal/internal/cache"
	"github.com/anvh2/futures-signal/internal/cache/basic"
	"github.com/anvh2/futures-signal/internal/cache/exchange"
	"github.com/anvh2/futures-signal/internal/cache/market"
	"github.com/anvh2/futures-signal/internal/logger"
	"github.com/anvh2/futures-signal/internal/services/binance"
	"github.com/anvh2/futures-signal/internal/services/telegram"
	"github.com/anvh2/futures-signal/internal/worker"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	blacklist = map[string]bool{}
)

type Server struct {
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

func New() *Server {
	logger, err := logger.New(viper.GetString("signaler.log_path"))
	if err != nil {
		log.Fatal("failed to init logger", err)
	}

	notify, err := telegram.NewTelegramBot(logger, viper.GetString("telegram.token"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	worker, err := worker.New(logger, &worker.PoolConfig{NumProcess: 8})
	if err != nil {
		log.Fatal("failed to new worker")
	}

	return &Server{
		logger:        logger,
		binance:       binance.New(logger),
		notify:        notify,
		worker:        worker,
		cache:         basic.NewCache(),
		marketCache:   market.NewMarket(viper.GetInt32("chart.candles.limit")),
		exchangeCache: exchange.New(logger),
		retryChannel:  make(chan *retry, 1000),
		quitChannel:   make(chan struct{}),
	}
}

func (s *Server) Start() error {
	s.worker.WithProcess(s.analyzing).Start()

	err := s.crawling()
	if err != nil {
		log.Fatal("failed to crawling data", zap.Error(err))
	}

	s.retrying()
	s.consuming()
	s.producing()
	s.notifying()

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		close(s.quitChannel)
		s.worker.Stop()

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
