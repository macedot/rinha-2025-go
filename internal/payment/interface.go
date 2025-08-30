package payment

import (
	"context"

	"github.com/macedot/rinha-2025-go/internal/types"
)

type writer interface {
	SavePayment(ctx context.Context, payload *types.PaymentRequest, processor int)
	Purge(ctx context.Context)
}

type reader interface {
	GetSummary(ctx context.Context, from, to string) types.SummaryResponse
}

type Repository interface {
	writer
	reader
}

type ProcessPayment func(ctx context.Context, data []byte)
type GetSummary func(ctx context.Context, from string, to string) types.SummaryResponse
type PurgePayments func(ctx context.Context)
