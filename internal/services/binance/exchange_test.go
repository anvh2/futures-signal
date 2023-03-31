package binance

import (
	"os"
	"testing"

	"github.com/anvh2/futures-signal/internal/logger"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	test_binanceInst *Binance
)

func TestMain(m *testing.M) {
	viper.SetDefault("binance.config.feed_url", "https://testnet.binancefuture.com")

	godotenv.Load("../../../.env")

	test_binanceInst = New(logger.NewDev())

	os.Exit(m.Run())
}
