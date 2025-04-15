package cache

import (
	"badminton-backend/internal/domain"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

// ErrKeyNotExist  Redis 特有的错误，用于表示键不存在
var ErrKeyNotExist = redis.Nil

// UserCache 用户缓存的接口
type UserCache interface {
	Get(ctx context.Context, id int64) (domain.User, error)
	Set(ctx context.Context, u domain.User) error
	Delete(ctx context.Context, id int64) error
}

// RedisUserCache 实现 UserCache 接口，封装了对 Redis 的操作
type RedisUserCache struct {
	cmd        redis.Cmdable
	expiration time.Duration
}

func NewRedisUserCache(cmd redis.Cmdable) UserCache {
	return &RedisUserCache{
		cmd:        cmd,
		expiration: time.Minute * 15, // 设置缓存过期时间为 15 分钟
	}
}

func (cache *RedisUserCache) Delete(ctx context.Context, id int64) error {
	return cache.cmd.Del(ctx, cache.key(id)).Err()
}

func (cache *RedisUserCache) Get(ctx context.Context, id int64) (domain.User, error) {
	key := cache.key(id)
	data, err := cache.cmd.Get(ctx, key).Result()
	if err != nil {
		return domain.User{}, err
	}
	var u domain.User
	err = json.Unmarshal([]byte(data), &u)
	return u, err
}

func (cache *RedisUserCache) Set(ctx context.Context, u domain.User) error {
	data, err := json.Marshal(u)
	if err != nil {
		return err
	}
	key := cache.key(u.Id)
	return cache.cmd.Set(ctx, key, data, cache.expiration).Err()
}

// key 生成缓存中用户数据的键
// 键的格式为 "user:info:{id}"，以便区分不同的用户
func (cache *RedisUserCache) key(id int64) string {
	return fmt.Sprintf("user:info:%d", id)
}
