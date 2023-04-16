package handler

import (
	"context"

	"github.com/anvh2/futures-signal/pkg/api/v1/signal"
)

func (h *Handler) ChangeTradingSettings(context.Context, *signal.ChangeTradingSettingsRequest) (*signal.ChangeTradingSettingsResponse, error) {
	return &signal.ChangeTradingSettingsResponse{}, nil
}
