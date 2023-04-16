package jobs

import (
	"context"
	"runtime/debug"
	"time"

	"github.com/anvh2/futures-signal/internal/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type retry struct {
	symbol   string
	interval string
	counter  *int
}

func (s *Jobs) Retrying() {
	for i := 0; i < 4; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					s.logger.Error("[Retry] failed to retry", zap.Any("error", r), zap.String("stacktrace", string(debug.Stack())))
				}
			}()

			for {
				select {
				case symbol := <-s.retryChannel:
					if symbol.counter == nil {
						symbol.counter = new(int)
					}

					delay(symbol.counter)

					resp, err := s.binance.ListCandlesticks(context.Background(), symbol.symbol, symbol.interval, viper.GetInt("chart.candles.limit"), 0, 0)
					if err != nil {
						s.logger.Error("[Retry] failed to get klines data", zap.String("symbol", symbol.symbol), zap.String("interval", symbol.interval), zap.Error(err))
						s.retryChannel <- symbol
						continue
					}

					for _, e := range resp {
						candle := &models.Candlestick{
							OpenTime:  e.OpenTime,
							CloseTime: e.CloseTime,
							Low:       e.Low,
							High:      e.High,
							Close:     e.Close,
						}

						s.marketCache.UpdateSummary(symbol.symbol).CreateCandle(symbol.interval, candle)
					}

					s.logger.Info("[Retry] success", zap.String("symbol", symbol.symbol), zap.String("interval", symbol.interval), zap.Int("total", len(resp)))

				case <-s.quitChannel:
					return
				}
			}
		}()
	}
}

func delay(counter *int) {
	*counter++
	if *counter%9 == 0 {
		time.Sleep(time.Minute)
	}

	duration := time.Duration(*counter * 100)
	time.Sleep(duration * time.Millisecond)
}
