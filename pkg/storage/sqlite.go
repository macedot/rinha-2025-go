package storage

import (
	"fmt"
	"log"
	"sync"

	"github.com/cvilsmeier/sqinn-go/v2"
	"github.com/macedot/rinha-2025-go/internal/types"
)

type SQLiteClient struct {
	sqinn *sqinn.Sqinn
}

func NewSQLiteClient() *SQLiteClient {
	sq := sqinn.MustLaunch(sqinn.Options{
		Db: ":memory:",
	})
	sq.MustExecSql("PRAGMA busy_timeout=5000")
	sq.MustExecSql("PRAGMA journal_mode=DELETE")
	sq.MustExecSql("PRAGMA synchronous=FULL")
	sq.MustExecSql("CREATE TABLE IF NOT EXISTS default_payments (amount INTEGER NOT NULL,ts TIMESTAMP NOT NULL)")
	sq.MustExecSql("CREATE TABLE IF NOT EXISTS fallback_payments (amount INTEGER NOT NULL,ts TIMESTAMP NOT NULL)")
	sq.MustExecSql("CREATE INDEX IF NOT EXISTS default_payments_ts_idx ON default_payments (ts)")
	sq.MustExecSql("CREATE INDEX IF NOT EXISTS fallback_payments_ts_idx ON fallback_payments (ts)")
	return &SQLiteClient{sq}
}

func (s *SQLiteClient) Close() {
	s.sqinn.Close()
}

func (s *SQLiteClient) InsertDefault(amount int64, ts string) error {
	err := s.sqinn.ExecParams("INSERT INTO default_payments (amount, ts) VALUES (?, ?)", 1, 2, []sqinn.Value{
		sqinn.Int64Value(1), sqinn.StringValue(ts),
	})
	return err
}

func (s *SQLiteClient) InsertFallback(amount int64, ts string) error {
	err := s.sqinn.ExecParams("INSERT INTO fallback_payments (amount, ts) VALUES (?, ?)", 1, 2,
		[]sqinn.Value{sqinn.Int64Value(1), sqinn.StringValue(ts)},
	)
	return err
}

func (s *SQLiteClient) GetSummary(from, to string) *types.SummaryResponse {
	summary := types.SummaryResponse{}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		summary.Default = *s.getTableSummary("default_payments", from, to)
	}()
	go func() {
		defer wg.Done()
		summary.Fallback = *s.getTableSummary("fallback_payments", from, to)
	}()
	wg.Wait()
	return &summary
}

func (s *SQLiteClient) getTableSummary(table string, from, to string) *types.SummaryServer {
	query, params := s.getSummaryParam(table, from, to)
	values, err := s.sqinn.QueryRows(
		query,
		params,                                 // query parameters
		[]byte{sqinn.ValInt32, sqinn.ValInt64}, // fetch id as int, name as string
	)
	if err != nil {
		log.Fatalln()
	}
	return &types.SummaryServer{
		TotalRequests: int(values[0][0].Int32),
		TotalAmount:   (values[0][1].Double / 100),
	}
}

func (s *SQLiteClient) getSummaryParam(table, from, to string) (string, []sqinn.Value) {
	if from != "" && to != "" {
		return fmt.Sprintf("SELECT COUNT(*),COLASENSE(SUM(amount), 0) FROM %s WHERE ts BETWEEN ? AND ?", table),
			[]sqinn.Value{sqinn.StringValue(from), sqinn.StringValue(to)}
	}
	return fmt.Sprintf("SELECT COUNT(*),COLASENSE(SUM(amount), 0) FROM %s", table), []sqinn.Value{}
}
