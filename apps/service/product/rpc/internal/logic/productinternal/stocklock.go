package productinternallogic

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// stockLockTTL 分布式锁TTL（防止死锁）
	stockLockTTL = 2 * time.Second
	// stockLockRetry 获取锁最大重试次数
	stockLockRetry = 3
	// stockLockRetryInterval 重试间隔
	stockLockRetryInterval = 50 * time.Millisecond
	// stockLockReleaseTimeout 释放锁的超时时间
	stockLockReleaseTimeout = 2 * time.Second
)

// releaseStockLockScript 释放锁：仅当锁的 value 与自己的 token 一致时才删除。
// 若不做这个比对（例如固定 value "1"），会出现：A 的锁 TTL 到期 → B 拿到锁 →
// A 的 defer 把 B 的锁删掉 → C 趁虚而入，最终两个请求并发写同一行库存。
var releaseStockLockScript = redis.NewScript(`
	if redis.call("GET", KEYS[1]) == ARGV[1] then
		return redis.call("DEL", KEYS[1])
	end
	return 0
`)

// stockLockToken 一次批量加锁的持有凭证，用于释放时校验归属
type stockLockToken struct {
	keys  []string
	value string
}

// stockLockKey 生成 SKU 库存操作的分布式锁 key
func stockLockKey(skuId uint64) string {
	return fmt.Sprintf("lock:stock:%d", skuId)
}

// acquireStockLocks 批量获取 SKU 分布式锁，全部成功时返回持有凭证。
//
// skuIds 会先排序去重：
//   - 排序保证所有请求按相同顺序加锁，与 DB 侧的行锁顺序一致，消除交叉顺序导致的死锁与互相抢锁；
//   - 去重避免同一个 SKU 重复出现时，第二次 SetNX 同一 key 必然失败导致整体加锁失败。
func acquireStockLocks(ctx context.Context, rdb *redis.Client, skuIds []uint64) (*stockLockToken, bool) {
	ids := normalizeSkuIds(skuIds)
	if len(ids) == 0 {
		return &stockLockToken{}, true
	}

	for retry := 0; retry < stockLockRetry; retry++ {
		token := &stockLockToken{
			keys:  make([]string, 0, len(ids)),
			value: uuid.New().String(),
		}
		allSuccess := true

		for _, skuId := range ids {
			key := stockLockKey(skuId)
			ok, err := rdb.SetNX(ctx, key, token.value, stockLockTTL).Result()
			if err != nil || !ok {
				allSuccess = false
				break
			}
			token.keys = append(token.keys, key)
		}

		if allSuccess {
			return token, true
		}

		// 加锁失败，释放本次已获取的部分后重试
		token.release(ctx, rdb)

		if retry < stockLockRetry-1 {
			time.Sleep(stockLockRetryInterval)
		}
	}
	return nil, false
}

// release 释放锁。基于调用方 context 派生一个「脱离取消」的独立 context：
//   - WithoutCancel 保留链路追踪等 value，但丢弃取消信号——调用方通常在 defer 中释放，
//     此时请求 context 可能已因超时/取消而失效；若直接复用，go-redis 在 pool.Get 阶段
//     就会因 ctx.Done() 提前返回，命令根本不发出、删除失败，锁只能等 TTL 自然过期；
//   - WithTimeout 给清理动作一个上界。释放失败不影响正确性（TTL 才是兜底），所以没有理由
//     在 Redis 故障时让请求同步阻塞数十秒。
func (t *stockLockToken) release(ctx context.Context, rdb *redis.Client) {
	if t == nil || len(t.keys) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stockLockReleaseTimeout)
	defer cancel()

	for _, key := range t.keys {
		if err := releaseStockLockScript.Run(ctx, rdb, []string{key}, t.value).Err(); err != nil {
			logx.WithContext(ctx).Errorf("释放库存锁失败, key=%s, error=%v", key, err)
		}
	}
}

// normalizeSkuIds 返回升序去重后的 SKU ID 列表，不修改入参
func normalizeSkuIds(skuIds []uint64) []uint64 {
	if len(skuIds) == 0 {
		return nil
	}
	ids := make([]uint64, len(skuIds))
	copy(ids, skuIds)
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	unique := ids[:1]
	for _, id := range ids[1:] {
		if id != unique[len(unique)-1] {
			unique = append(unique, id)
		}
	}
	return unique
}
