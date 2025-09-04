package database

// Thanks andersongomes001 for the great idea to use in Redis:
// https://github.com/andersongomes001/rinha-2025-go/blob/health/src/infrastructure/redis.rs

import (
	"context"
	"rinha-2025-go/internal/config"
	"rinha-2025-go/internal/models"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	ctx context.Context
	rdb *redis.Client
}

func NewRedisClient(cfg *config.Config) *Redis {
	return &Redis{
		ctx: context.Background(),
		rdb: redis.NewClient(&redis.Options{
			Addr: cfg.RedisSocket,
		}),
	}
}

func (r *Redis) Close() {
	r.rdb.Close()
}

func (r *Redis) SavePayment(instance *config.Service, payment *models.Payment) error {
	//amount := strconv.Itoa(int(math.Round(payment.Amount * 100)))
	ts := float64(payment.Timestamp.UnixNano()) / 1e9
	if err := r.rdb.HSet(r.ctx, instance.KeyAmount, payment.PaymentID, payment.Amount).Err(); err != nil {
		return err
	}
	if err := r.rdb.ZAdd(r.ctx, instance.KeyTime, redis.Z{Score: ts, Member: payment.PaymentID}).Err(); err != nil {
		return err
	}
	return nil
}

func (r *Redis) RemovePayment(instance *config.Service, payment *models.Payment) error {
	return r.rdb.HDel(r.ctx, instance.KeyTime, payment.PaymentID).Err()
}

func (r *Redis) GetSummary(instance *config.Service, summary *models.SummaryParam) (*models.ProcessorSummary, error) {
	var res models.ProcessorSummary
	ids, err := r.rdb.ZRangeByScore(r.ctx, instance.KeyTime,
		&redis.ZRangeBy{Min: summary.StartTime, Max: summary.EndTime}).Result()
	if err != nil {
		return nil, err
	}
	if len(ids) > 0 {
		amounts, err := r.rdb.HMGet(r.ctx, instance.KeyAmount, ids...).Result()
		if err != nil {
			return nil, err
		}
		res.RequestCount = len(amounts)
		for _, val := range amounts {
			// if i, err := strconv.ParseFloat(val.(string), 32); err == nil {
			// 	res.TotalAmount += i
			// }
			i, _ := strconv.ParseFloat(val.(string), 32)
			res.TotalAmount += i
		}
	}
	return &res, nil
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
