package base62

import (
	"fmt"
)

// Base62 字符集定义 (62个字符: 0-9, A-Z, a-z)
const base62Charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
const base = 62

// base62Map 用于高效地将 Base62 字符映射回 uint64 数值 (0-61)
var base62Map = map[rune]uint64{}

// init 函数会在包被导入时自动运行一次，用于初始化查找表
func init() {
	for i, char := range base62Charset {
		base62Map[char] = uint64(i)
	}
}

// reverse 反转 []byte 切片。用于 Encode 函数修正编码顺序。
func reverse(b []byte) []byte {
	for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
		b[i], b[j] = b[j], b[i]
	}
	return b
}

// Encode 将一个 uint64 ID (如雪花ID) 转换为 Base62 编码的字符串。
// 编码逻辑：短除法，先追加低位，后反转。
func Encode(seq uint64) string {
	if seq == 0 {
		return string(base62Charset[0])
	}

	bl := []byte{}

	for seq > 0 {
		mod := seq % base
		// 追加低位字符
		bl = append(bl, base62Charset[mod])
		seq /= base
	}

	// 反转，使编码顺序从高位到低位
	return string(reverse(bl))
}

// Decode 将 Base62 编码的字符串反解码为 uint64 ID。
// 反解码逻辑：按权展开求和。
func Decode(base62 string) (uint64, error) {
	var id uint64 = 0
	// 从左到右遍历 Base62 字符串
	for _, char := range base62 {
		val, exists := base62Map[char]
		if !exists {
			// 如果发现非 Base62 字符，则返回错误
			return 0, fmt.Errorf("invalid Base62 character: %c", char)
		}
		// 按权展开求和: id = id * Base + val
		id = id*base + val
	}

	return id, nil
}

// GenerateNickname 生成带前缀的唯一昵称
func GenerateNickname(prefix string, snowflakeId uint64) string {
	// 雪花ID → Base62短串
	shortStr := Encode(snowflakeId)
	// 前缀 + 短串 = 最终昵称
	return prefix + shortStr
}
