package database

// Thanks andersongomes001 for the great idea to use in Redis:
// https://github.com/andersongomes001/rinha-2025/blob/health/src/infrastructure/redis.rs

import (
	"context"
	"fmt"
	"math"
	"rinha-2025/config"
	"rinha-2025/models"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	rdb *redis.Client
}

var redisInstance Redis

func RedisInstance() *Redis {
	return &redisInstance
}

func (r *Redis) Connect(cfg *config.Config) *Redis {
	if r.rdb != nil {
		r.Close()
	}
	r.rdb = redis.NewClient(&redis.Options{
		Addr: cfg.RedisURL,
	})
	return r
}

func (r *Redis) Close() {
	r.rdb.Close()
}

func (r *Redis) SavePayment(instance *config.Service, payment *models.Payment) error {
	amount := strconv.Itoa(int(math.Round(payment.Amount * 100)))
	ts := float64(payment.Timestamp.UnixNano()) / 1e9
	ctx := context.Background()
	if err := r.rdb.HSet(ctx, instance.KeyAmount, payment.PaymentID, amount).Err(); err != nil {
		return err
	}
	if err := r.rdb.ZAdd(ctx, instance.KeyTime, redis.Z{Score: ts, Member: payment.PaymentID}).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Redis) GetSummary(instance *config.Service, summary *models.SummaryParam) (models.ProcessorSummary, error) {
	var res models.ProcessorSummary
	ctx := context.Background()
	ids, err := r.rdb.ZRangeByScore(ctx, instance.KeyTime,
		&redis.ZRangeBy{Min: summary.StartTime, Max: summary.EndTime}).Result()
	if err != nil {
		return res, err
	}
	res.RequestCount = len(ids)
	if res.RequestCount > 0 {
		amounts, err := r.rdb.HMGet(ctx, instance.KeyAmount, ids...).Result()
		if err != nil {
			return res, err
		}
		if res.RequestCount != len(amounts) {
			return res, fmt.Errorf("%d != %d", res.RequestCount, len(amounts))
		}
		total := int64(0)
		for _, val := range amounts {
			if i, err := strconv.ParseInt(val.(string), 10, 64); err == nil {
				total += i
			}
		}
		res.TotalAmount = float64(total) / 100
	}
	return res, nil
}

func (r *Redis) FlushAll() error {
	ctx := context.Background()
	r.rdb.FlushDB(ctx)
	return nil
}

func (r *Redis) SetString(key, label, value string) error {
	ctx := context.Background()
	return r.rdb.HSet(ctx, key, label, value).Err()
}

func (r *Redis) GetString(key, label string) string {
	ctx := context.Background()
	str, err := r.rdb.HGet(ctx, key, label).Result()
	if err != nil {
		str = ""
	}
	return str
}

func (r *Redis) SetInt(key, label string, value int64) error {
	return r.SetString(key, label, strconv.FormatInt(value, 10))
}

func (r *Redis) GetInt(key, label string) int64 {
	str := r.GetString(key, label)
	num, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		num = 0
	}
	return num
}
