package cache

import (
	"sync"
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
)

const timeTruncate = time.Second

type TimeBucket struct {
	Records []*types.PaymentRecord
	Total   int64
}

type PaymentStore struct {
	mu      sync.Mutex
	buckets map[int64]*TimeBucket // Key: Unix time (seconds)
}

func NewStore() *PaymentStore {
	return &PaymentStore{
		buckets: make(map[int64]*TimeBucket),
	}
}

func (db *PaymentStore) AddRecord(record *types.PaymentRecord) {
	db.mu.Lock()
	defer db.mu.Unlock()
	bucketKey := record.RequestedAt.Truncate(timeTruncate).Unix()
	bucket := db.buckets[bucketKey]
	bucket.Records = append(db.buckets[bucketKey].Records, record)
	bucket.Total += record.Amount
}

func (db *PaymentStore) QuerySummary(from, to time.Time) types.SummaryServer {
	db.mu.Lock()
	defer db.mu.Unlock()
	var summary struct {
		TotalRequests int
		TotalAmount   int64
	}
	fromT := from.Truncate(timeTruncate)
	toT := to.Add(timeTruncate).Truncate(timeTruncate)
	currentTime := fromT
	for !currentTime.After(toT) {
		bucket := db.buckets[currentTime.Unix()]
		nextTime := currentTime.Add(timeTruncate)
		if (!currentTime.Before(from)) &&
			(nextTime.Before(to) || nextTime.Equal(to)) {
			summary.TotalRequests += len(bucket.Records)
			summary.TotalAmount += bucket.Total
		} else {
			for _, record := range bucket.Records {
				if record.RequestedAt.Before(from) {
					continue
				}
				if !record.RequestedAt.Before(to) {
					continue
				}
				summary.TotalRequests++
				summary.TotalAmount += record.Amount
			}
		}
		currentTime = nextTime
	}
	return types.SummaryServer{
		TotalRequests: summary.TotalRequests,
		TotalAmount:   float64(summary.TotalAmount) / 100,
	}
}

func (db *PaymentStore) Clean() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.buckets = make(map[int64]*TimeBucket)
}

type MemoryDB struct {
	Default  *PaymentStore
	Fallback *PaymentStore
}

func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		Default:  NewStore(),
		Fallback: NewStore(),
	}
}

func (db *MemoryDB) AddRecordDefault(record *types.PaymentRecord) {
	db.Default.AddRecord(record)
}

func (db *MemoryDB) AddRecordFallback(record *types.PaymentRecord) {
	db.Fallback.AddRecord(record)
}

func (db *MemoryDB) QuerySummary(from, to time.Time) types.SummaryResponse {
	var wg sync.WaitGroup
	var summary types.SummaryResponse
	wg.Add(2)
	go func() {
		defer wg.Done()
		summary.Default = db.Default.QuerySummary(from, to)
	}()
	go func() {
		defer wg.Done()
		summary.Fallback = db.Fallback.QuerySummary(from, to)
	}()
	wg.Wait()
	return summary
}

func (db *MemoryDB) Clean() {
	db.Default.Clean()
	db.Fallback.Clean()
}
