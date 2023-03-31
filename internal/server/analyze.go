package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/anvh2/futures-signal/internal/helpers"
	"github.com/anvh2/futures-signal/internal/models"
	"github.com/anvh2/futures-signal/internal/talib"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func (s *Server) analyzing(ctx context.Context, data interface{}) error {
	message := &models.CandleChart{
		Candles:  make(map[string][]*models.Candlestick),
		Metadata: make(map[string]*models.ChartMetadata),
	}

	if err := json.Unmarshal([]byte(fmt.Sprint(data)), message); err != nil {
		s.logger.Error("[Process] failed to unmarshal message", zap.Error(err))
		return err
	}

	if err := validateMessage(message); err != nil {
		s.logger.Error("[Process] failed to validate message", zap.Error(err))
		return err
	}

	oscillator := &models.Oscillator{
		Symbol: message.Symbol,
		Stoch:  make(map[string]*models.Stoch),
	}

	for interval, candles := range message.Candles {
		low := make([]float64, len(candles))
		high := make([]float64, len(candles))
		close := make([]float64, len(candles))

		for idx, candle := range candles {
			l, _ := strconv.ParseFloat(candle.Low, 64)
			low[idx] = l

			h, _ := strconv.ParseFloat(candle.High, 64)
			high[idx] = h

			c, _ := strconv.ParseFloat(candle.Close, 64)
			close[idx] = c
		}

		_, rsi := talib.RSIPeriod(14, close)
		k, d, _ := talib.KDJ(9, 3, 3, high, low, close)

		stoch := &models.Stoch{
			RSI: rsi[len(rsi)-1],
			K:   k[len(k)-1],
			D:   d[len(d)-1],
		}

		oscillator.Stoch[interval] = stoch

		if !talib.WithinRangeBound(oscillator.Stoch[interval], talib.RangeBoundRecommend) {
			return errors.New("analyze: not ready to trade")
		}

		msg := fmt.Sprintf("%s\t\t\t latest: -%0.4f(s)\n\t%s\n", message.Symbol, float64((time.Now().UnixMilli()-message.Metadata[interval].UpdateTime))/1000.0, helpers.ResolvePositionSide(oscillator.GetRSI()))

		for _, interval := range viper.GetStringSlice("market.intervals") {
			stoch, ok := oscillator.Stoch[interval]
			if !ok {
				s.logger.Error("[Process] stoch in interval invalid", zap.Any("stoch", stoch))
				return errors.New("analyze: stoch in interval invalid")
			}

			msg += fmt.Sprintf("\t%03s:\t RSI %2.2f | K %02.2f | D %02.2f\n", strings.ToUpper(interval), stoch.RSI, stoch.K, stoch.D)
		}

		upserted, err := s.redisCli.SetNX(ctx, fmt.Sprintf("signaler.sent.%s-%s", message.Symbol, interval), true, 10*time.Minute).Result()
		if err != nil || !upserted {
			return err
		}

		err = s.notify.PushNotify(ctx, viper.GetInt64("notify.channels.futures_recommendation"), msg)
		if err != nil {
			s.logger.Error("[Process] failed to push notification", zap.Error(err))
			return err
		}
	}

	return nil
}

func validateMessage(message *models.CandleChart) error {
	if message == nil {
		return errors.New("notify: message invalid")
	}
	return nil
}
