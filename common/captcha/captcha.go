package captcha

import (
	"context"
	"crypto/rand"
	"fmt"
	"github.com/ZY0506/PrimeMall/common/constants"
	"github.com/ZY0506/PrimeMall/common/response"
	"github.com/go-redis/redis/v8"
	"github.com/zeromicro/go-zero/core/logx"
	"math/big"
)

type Service struct {
	rdb *redis.Client
}

func NewService(rdb *redis.Client) *Service {
	return &Service{
		rdb: rdb,
	}
}

// Send 发送验证码（频率校验 + 生成 + 存储 + 发送回调）
func (s *Service) Send(ctx context.Context, phone, scene string, sendFunc func(phone, code string) error) error {
	// 冷却检查、生成、存储、调用 sendFunc（暂时不实现）、失败回滚

	// 频率校验
	coolKey := fmt.Sprintf("%s%s:%s", constants.CAPTCHA_COOL_KEY, scene, phone)
	ok, err := s.rdb.SetNX(ctx, coolKey, "1", constants.CAPTCHA_COOL_DOWN).Result()
	if err != nil {
		return err
	}
	if !ok {
		return response.NewBizError(response.ErrCodeTooFrequent, "操作过于频繁")
	}

	// 生成验证码（crypto/rand）
	code, err := generateSecureCode(6)
	if err != nil {
		// 生成失败，删除冷却标记
		s.rdb.Del(ctx, coolKey)
		return err
	}

	// 存储验证码到 Redis
	storeKey := fmt.Sprintf("%s%s:%s", constants.CAPTCHA_KEY, scene, phone)
	err = s.rdb.Set(ctx, storeKey, code, constants.CAPTCHA_EXPIRE).Err()
	if err != nil {
		// 存储失败，删除冷却标记，避免锁死
		s.rdb.Del(ctx, coolKey)
		return err
	}

	// 模拟发送
	// TODO: 调用短信服务以及回调、失败回滚
	//logx.Infof("发送验证码给手机: %s", phone)
	logx.WithContext(ctx).Infof("发送验证码给手机: %s", phone)
	fmt.Printf("验证码: %s", string(code))
	// 调用发送回调
	if err := sendFunc(phone, code); err != nil {
		// 发送失败，回滚冷却标记和验证码
		s.rdb.Del(ctx, coolKey, storeKey)
		return err
	}

	return nil
}

// Verify 原子验证验证码（GET + DEL）
// 返回 true 表示验证成功且已删除
func (s *Service) Verify(ctx context.Context, phone, scene, inputCode string) (bool, error) {
	storeKey := fmt.Sprintf("%s%s:%s", constants.CAPTCHA_KEY, scene, phone)

	// Lua 脚本保证原子性
	const luaScript = `
        local key = KEYS[1]
        local input = ARGV[1]
        local stored = redis.call('GET', key)
        if stored == input then
            redis.call('DEL', key)
            return 1
        else
            return 0
        end
    `
	result, err := s.rdb.Eval(ctx, luaScript, []string{storeKey}, inputCode).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}

// 辅助函数：生成安全验证码
func generateSecureCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[n.Int64()]
	}
	return string(code), nil
}
