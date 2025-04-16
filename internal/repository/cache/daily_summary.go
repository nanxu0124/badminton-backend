package cache

import (
	"badminton-backend/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type DailySummaryCache interface {
	Get(ctx context.Context, biz string, userID int64, date time.Time) (domain.DailySummary, error)
	Set(ctx context.Context, biz string, userID int64, date time.Time, u domain.DailySummary) error
	Delete(ctx context.Context, biz string, userID int64, date time.Time) error
}

type RedisDailySummaryCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisDailySummaryCache(cmd redis.Cmdable) DailySummaryCache {
	return &RedisDailySummaryCache{
		cmd:        cmd,
		expiration: time.Minute * 15,
	}
}

func (cache *RedisDailySummaryCache) Get(ctx context.Context, biz string, userID int64, date time.Time) (domain.DailySummary, error) {
	key := cache.key(biz, userID, date)
	data, err := cache.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.DailySummary{}, err
	}
	var u domain.DailySummary
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (cache *RedisDailySummaryCache) Set(ctx context.Context, biz string, userID int64, date time.Time, u domain.DailySummary) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(biz, userID, date)
	return cache.cmd.Set(ctx, key, data, cache.expiration).Err()
}

func (cache *RedisDailySummaryCache) Delete(ctx context.Context, biz string, userID int64, date time.Time) error {
	return cache.cmd.Del(ctx, cache.key(biz, userID, date)).Err()
}

func (cache *RedisDailySummaryCache) key(biz string, userID int64, date time.Time) string {
	return fmt.Sprintf("DailySummary:info:%s-%d-%s", biz, userID, date.Format(time.DateOnly))
}
