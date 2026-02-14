package database

// Thanks andersongomes001 for the great idea to use in Redis:
// https://github.com/andersongomes001/rinha-2025-go/blob/health/src/infrastructure/redis.rs

import (
	"context"
	"fmt"
	"log"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/models"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	ctx context.Context
	Rdb *redis.Client
}

func NewRedisClient(cfg *config.Config) *Redis {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisSocket,
		PoolSize: 200,
	})
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}
	return &Redis{
		ctx: ctx,
		Rdb: rdb,
	}
}

func (r *Redis) Close() {
	r.Rdb.Close()
}

func (r *Redis) SavePayment(instance *config.Service, payment *models.Payment) error {
	ts := float64(payment.Timestamp.UnixNano()) / 1e9
	pipe := r.Rdb.Pipeline()
	pipe.HSet(r.ctx, instance.KeyAmount, payment.PaymentID, payment.Amount)
	pipe.ZAdd(r.ctx, instance.KeyTime, redis.Z{Score: ts, Member: payment.PaymentID})
	_, err := pipe.Exec(r.ctx)
	return err
}

func (r *Redis) RemovePayment(instance *config.Service, payment *models.Payment) error {
	return r.Rdb.HDel(r.ctx, instance.KeyTime, payment.PaymentID).Err()
}

func (r *Redis) GetSummary(instance *config.Service, summary *models.SummaryParam) *models.ProcessorSummary {
	var res models.ProcessorSummary
	ids, err := r.Rdb.ZRangeByScore(r.ctx, instance.KeyTime,
		&redis.ZRangeBy{Min: summary.StartTime, Max: summary.EndTime}).Result()
	if err != nil {
		log.Println("GetSummary:ZRangeByScore:", instance.Table, ":", err)
		return &res
	}
	if len(ids) > 0 {
		amounts, err := r.Rdb.HMGet(r.ctx, instance.KeyAmount, ids...).Result()
		if err != nil {
			log.Println("GetSummary:HMGet:", instance.Table, ":", err)
			return &res
		}
		res.RequestCount = len(amounts)
		for _, val := range amounts {
			i, _ := strconv.ParseFloat(val.(string), 32)
			res.TotalAmount += i
		}
	}
	return &res
}

func (r *Redis) FlushAll() error {
	ctx := context.Background()
	r.Rdb.FlushDB(ctx)
	return nil
}

func (r *Redis) SetString(key, label, value string) error {
	ctx := context.Background()
	return r.Rdb.HSet(ctx, key, label, value).Err()
}

func (r *Redis) GetString(key, label string) string {
	ctx := context.Background()
	str, err := r.Rdb.HGet(ctx, key, label).Result()
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

func (r *Redis) TryLock(lockKey string, lockValue string, ttl time.Duration) bool {
	success, err := r.Rdb.SetNX(r.ctx, lockKey, lockValue, ttl).Result()
	if err != nil {
		log.Printf("failed to acquire lock: %s", err.Error())
		return false
	}
	return success
}

func (r *Redis) Unlock(lockKey string) error {
	return r.Rdb.Del(r.ctx, lockKey).Err()
}

func (r *Redis) GetLastRunTime(timeKey string) (time.Time, error) {
	result, err := r.Rdb.Get(r.ctx, timeKey).Result()
	if err == redis.Nil {
		return time.Time{}, nil
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last run time: %w", err)
	}
	t, err := time.Parse(time.RFC3339Nano, result)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse last run time: %w", err)
	}
	return t, nil
}

func (r *Redis) SetLastRunTime(timeKey string, t time.Time) error {
	err := r.Rdb.Set(r.ctx, timeKey, t.Format(time.RFC3339Nano), 0).Err()
	if err != nil {
		return fmt.Errorf("failed to set last run time: %w", err)
	}
	return nil
}

func (r *Redis) ResetStat(key string) error {
	return r.Rdb.Set(r.ctx, key, "0", 0).Err()
}

func (r *Redis) AddStat(key string) error {
	return r.Rdb.Incr(r.ctx, key).Err()
}

func (r *Redis) GetStat(key string) int64 {
	result, err := r.Rdb.Get(r.ctx, key).Result()
	if err == redis.Nil || err != nil {
		return 0
	}
	value, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0
	}
	return value
}
