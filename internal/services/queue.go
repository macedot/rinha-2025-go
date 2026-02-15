package services

import (
	"context"
	"fmt"
	"rinha-2025-go/internal/database"
	"rinha-2025-go/internal/models"
	"time"

	"github.com/ohler55/ojg/oj"
	"github.com/redis/go-redis/v9"
)

type PaymentQueue struct {
	ctx    context.Context
	key    string // Redis key for the queue (list)
	client *redis.Client
}

func NewPaymentQueue(ctx context.Context, redis *database.Redis) *PaymentQueue {
	return &PaymentQueue{
		ctx:    ctx,
		key:    "payment-queue",
		client: redis.Rdb,
	}
}

func (q *PaymentQueue) Enqueue(payment *models.Payment) error {
	// Get buffer from pool for JSON marshaling
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)

	data, err := oj.Marshal(payment, *bufPtr)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	err = q.client.RPush(q.ctx, q.key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to push to queue: %w", err)
	}
	return nil
}

func (q *PaymentQueue) Dequeue() *models.Payment {
	result, err := q.client.BLPop(q.ctx, 1*time.Second, q.key).Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return nil
	}
	var payment models.Payment
	err = oj.Unmarshal([]byte(result[1]), &payment)
	if err != nil {
		return nil
	}
	return &payment
}

func (q *PaymentQueue) Length() int64 {
	length, err := q.client.LLen(q.ctx, q.key).Result()
	if err != nil {
		return 0
	}
	return length
}

func (q *PaymentQueue) Close() error {
	// Client is shared from database package, not closed here
	return nil
}
