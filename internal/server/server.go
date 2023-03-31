package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anvh2/futures-signal/internal/cache"
	"github.com/anvh2/futures-signal/internal/cache/exchange"
	"github.com/anvh2/futures-signal/internal/cache/market"
	"github.com/anvh2/futures-signal/internal/logger"
	"github.com/anvh2/futures-signal/internal/services/binance"
	"github.com/anvh2/futures-signal/internal/services/telegram"
	"github.com/anvh2/futures-signal/internal/worker"
	"github.com/go-redis/redis/v8"
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
	redisCli      *redis.Client
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

	redisCli := redis.NewClient(&redis.Options{
		Addr:       viper.GetString("redis.addr"),
		Password:   viper.GetString("redis.pass"),
		DB:         1,
		MaxRetries: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := redisCli.Ping(ctx).Err(); err != nil {
		log.Fatal("failed to connect to redis", err)
	}

	notify, err := telegram.NewTelegramBot(logger, viper.GetString("telegram.token"))
	if err != nil {
		log.Fatal("failed to new chat bot", err)
	}

	worker, err := worker.New(logger, &worker.PoolConfig{NumProcess: 8})
	if err != nil {
		log.Fatal("failed to new worker")
	}

	market := market.NewMarket(viper.GetInt32("chart.candles.limit"))

	exchange := exchange.New(logger)
	binance := binance.New(logger)

	return &Server{
		logger:        logger,
		binance:       binance,
		notify:        notify,
		worker:        worker,
		redisCli:      redisCli,
		marketCache:   market,
		exchangeCache: exchange,
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

	sigs := make(chan os.Signal, 1)
	done := make(chan bool)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("Server now listening")

	go func() {
		<-sigs
		close(s.quitChannel)
		s.worker.Stop()
		s.redisCli.Close()

		close(done)
	}()

	fmt.Println("Ctrl-C to interrupt...")
	<-done
	fmt.Println("Exiting...")

	return nil
}
