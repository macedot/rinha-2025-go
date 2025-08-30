package store

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/macedot/rinha-2025-go/internal/types"
	"github.com/macedot/rinha-2025-go/pkg/util"
)

type PaymentDB struct {
	conn *pgxpool.Pool
}

func NewPaymentDB(ctx context.Context) (*PaymentDB, error) {
	cfg, _ := pgxpool.ParseConfig(util.GetEnvOr("DATABASE_URL", "postgresql://rinha:rinha@postgres:5432/rinha"))
	maxConns, _ := strconv.Atoi(util.GetEnvOr("PG_MAX_CONNS", "25"))
	cfg.MaxConns = int32(maxConns)
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = time.Hour
	cfg.MinConns = 20
	conn, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return &PaymentDB{conn: conn}, nil
}

func (s *PaymentDB) Close() {
	s.conn.Close()
}

func (s *PaymentDB) SavePayment(ctx context.Context, payload *types.PaymentRequest, processor int) {
	_, err := s.conn.Exec(ctx, `
		INSERT INTO payments(correlation_id, amount, processor, requested_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (correlation_id) DO NOTHING;`,
		payload.CorrelationID, payload.Amount, processor, payload.RequestedAt)
	if err != nil {
		log.Printf("Erro na entrega do postgres: %v", err)
	}
}

func (s *PaymentDB) GetSummary(ctx context.Context, from, to string) types.SummaryResponse {
	response := types.SummaryResponse{}
	f, t := time.Unix(0, 0).UTC().Format(time.RFC3339Nano), time.Now().UTC().AddDate(0, 0, 1).Format(time.RFC3339Nano)
	if from != "" {
		f = from
	}
	if to != "" {
		t = to
	}
	rows, err := s.conn.Query(ctx, `
		SELECT processor, COUNT(*) AS total, SUM(amount)
		FROM payments
		WHERE requested_at BETWEEN $1 AND $2
		GROUP BY processor;`,
		f, t)
	if err != nil {
		return response
	}
	defer rows.Close()
	for rows.Next() {
		var processor int
		var total int
		var sum float64
		err := rows.Scan(&processor, &total, &sum)
		if err == nil {
			if processor == 1 {
				response.Default.TotalRequests = total
				response.Default.TotalAmount = sum
			} else {
				response.Fallback.TotalRequests = total
				response.Fallback.TotalAmount = sum
			}
		}
	}
	return response
}

func (s *PaymentDB) Purge(ctx context.Context) {
	s.conn.Exec(ctx, "TRUNCATE TABLE payments RESTART IDENTITY;")
}
