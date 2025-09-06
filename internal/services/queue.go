package services

import (
	"context"
	"fmt"
	"log"
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

func NewPaymentQueue(ctx context.Context, redisSocket string) *PaymentQueue {
	client := redis.NewClient(&redis.Options{
		Addr:     redisSocket,
		PoolSize: 200,
	})
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Printf("failed to connect to redis: %s", err.Error())
		return nil
	}
	return &PaymentQueue{
		ctx:    ctx,
		key:    "payment-queue",
		client: client,
	}
}

func (q *PaymentQueue) Enqueue(payment *models.Payment) error {
	data, err := oj.Marshal(payment)
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
		log.Printf("failed to pop from queue: %s", err.Error())
		return nil
	}
	var payment models.Payment
	err = oj.Unmarshal([]byte(result[1]), &payment)
	if err != nil {
		log.Printf("failed to unmarshal message: %s", err.Error())
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
	return q.client.Close()
}
