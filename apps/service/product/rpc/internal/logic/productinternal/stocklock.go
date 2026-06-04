package productinternallogic

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// stockLockTTL 分布式锁TTL（防止死锁）
	stockLockTTL = 2 * time.Second
	// stockLockRetry 获取锁最大重试次数
	stockLockRetry = 3
	// stockLockRetryInterval 重试间隔
	stockLockRetryInterval = 50 * time.Millisecond
)

// stockLockKey 生成 SKU 库存操作的分布式锁 key
func stockLockKey(skuId uint64) string {
	return fmt.Sprintf("lock:stock:%d", skuId)
}

// acquireStockLocks 批量获取 SKU 分布式锁，全部成功返回 true
func acquireStockLocks(ctx context.Context, rdb *redis.Client, skuIds []uint64) bool {
	for retry := 0; retry < stockLockRetry; retry++ {
		acquired := make([]string, 0, len(skuIds))
		allSuccess := true

		for _, skuId := range skuIds {
			key := stockLockKey(skuId)
			ok, err := rdb.SetNX(ctx, key, "1", stockLockTTL).Result()
			if err != nil || !ok {
				allSuccess = false
				break
			}
			acquired = append(acquired, key)
		}

		if allSuccess {
			return true
		}

		releaseStockLocks(ctx, rdb, acquired)

		if retry < stockLockRetry-1 {
			time.Sleep(stockLockRetryInterval)
		}
	}
	return false
}

// releaseStockLocks 释放 SKU 分布式锁
func releaseStockLocks(ctx context.Context, rdb *redis.Client, keys []string) {
	if len(keys) == 0 {
		return
	}
	// Lua脚本：原子删除，确保只删除自己持有的锁
	script := `
		if redis.call("GET", KEYS[1]) == "1" then
			return redis.call("DEL", KEYS[1])
		end
		return 0
	`
	for _, key := range keys {
		_ = rdb.Eval(ctx, script, []string{key}).Err()
	}
}

// releaseStockLocksByIds 根据 SKU ID 列表释放分布式锁
func releaseStockLocksByIds(ctx context.Context, rdb *redis.Client, skuIds []uint64) {
	if len(skuIds) == 0 {
		return
	}
	keys := make([]string, len(skuIds))
	for i, id := range skuIds {
		keys[i] = stockLockKey(id)
	}
	releaseStockLocks(ctx, rdb, keys)
}
