package handler

import (
	"context"

	"github.com/anvh2/futures-signal/pkg/api/v1/signal"
)

func (h *Handler) PerformSignalAnalysis(context.Context, *signal.PerformSignalRequestAnalysis) (*signal.PerformSignalResponseAnalysis, error) {
	return &signal.PerformSignalResponseAnalysis{}, nil
}
