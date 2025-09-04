package store

import (
	"sort"
	"sync"
	"time"

	"github.com/macedot/rinha-2025-go/internal/types"
)

const timeTruncate = time.Second

type ServerSum struct {
	Count int
	Sum   int64
}

type TimeBucket struct {
	StartTime time.Time // Start of the millisecond (truncated)
	Records   []*types.PaymentRecord
	ServerSum ServerSum
}

type PaymentStore struct {
	mu        sync.Mutex
	buckets   map[int64]*TimeBucket // Key: UnixNano of millisecond
	timeIndex []*TimeBucket         // Sorted slice of buckets
}

func NewStore() *PaymentStore {
	return &PaymentStore{
		buckets:   make(map[int64]*TimeBucket),
		timeIndex: make([]*TimeBucket, 0, 1024*1024),
	}
}

func (db *PaymentStore) insertToTimeIndex(bucket *TimeBucket) {
	index := sort.Search(len(db.timeIndex), func(i int) bool {
		return !db.timeIndex[i].StartTime.Before(bucket.StartTime)
	})
	db.timeIndex = append(db.timeIndex, nil)
	copy(db.timeIndex[index+1:], db.timeIndex[index:])
	db.timeIndex[index] = bucket
}

func (db *PaymentStore) AddRecord(record *types.PaymentRecord) {
	db.mu.Lock()
	defer db.mu.Unlock()

	truncated := record.RequestedAt.Truncate(timeTruncate)
	key := truncated.UnixNano()

	bucket, exists := db.buckets[key]
	if !exists {
		bucket = &TimeBucket{
			StartTime: truncated,
			Records:   make([]*types.PaymentRecord, 0, 100),
			ServerSum: ServerSum{},
		}
		db.buckets[key] = bucket
		db.insertToTimeIndex(bucket)
	}

	bucket.Records = append(bucket.Records, record)
	bucket.ServerSum.Count++
	bucket.ServerSum.Sum += record.Amount
}

func (db *PaymentStore) QuerySummary(from, to time.Time) types.SummaryServer {
	db.mu.Lock()
	defer db.mu.Unlock()

	// Initialize temporary result storage
	var tempResult ServerSum

	// Find bucket range boundaries with millisecond precision
	fromMS := from.Truncate(timeTruncate)
	toMS := to.Truncate(timeTruncate)

	// Find starting index using binary search
	startIdx := sort.Search(len(db.timeIndex), func(i int) bool {
		return !db.timeIndex[i].StartTime.Before(fromMS)
	})

	// Process buckets in time range
	for i := startIdx; i < len(db.timeIndex); i++ {
		bucket := db.timeIndex[i]
		bucketEnd := bucket.StartTime.Add(timeTruncate)
		// Stop if we've passed the end of our time range
		if !to.IsZero() && bucket.StartTime.After(toMS) {
			break
		}
		// Check if bucket is entirely within query range
		if (from.IsZero() || !bucket.StartTime.Before(from)) &&
			(to.IsZero() || bucketEnd.Before(to) || bucketEnd.Equal(to)) {
			// Use pre-aggregated data for full bucket
			tempResult.Count += bucket.ServerSum.Count
			tempResult.Sum += bucket.ServerSum.Sum
		} else {
			// Partial bucket - check individual records
			for _, record := range bucket.Records {
				if !from.IsZero() && record.RequestedAt.Before(from) {
					continue
				}
				if !to.IsZero() && !record.RequestedAt.Before(to) {
					continue
				}
				tempResult.Count++
				tempResult.Sum += record.Amount
			}
		}
	}

	return types.SummaryServer{
		TotalRequests: tempResult.Count,
		TotalAmount:   float64(tempResult.Sum) / 100,
	}
}

func (db *PaymentStore) Clean() {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.buckets = make(map[int64]*TimeBucket)
	db.timeIndex = make([]*TimeBucket, 0, 1024*1024)
}

type PaymentDB struct {
	Default  *PaymentStore
	Fallback *PaymentStore
}

func NewPaymentDB() *PaymentDB {
	return &PaymentDB{
		Default:  NewStore(),
		Fallback: NewStore(),
	}
}

func (db *PaymentDB) AddRecordDefault(record *types.PaymentRecord) {
	db.Default.AddRecord(record)
}

func (db *PaymentDB) AddRecordFallback(record *types.PaymentRecord) {
	db.Fallback.AddRecord(record)
}

func (db *PaymentDB) QuerySummary(from, to time.Time) types.SummaryResponse {
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

func (db *PaymentDB) Clean() {
	db.Default.Clean()
	db.Fallback.Clean()
}
