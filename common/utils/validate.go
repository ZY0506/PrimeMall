package utils

import (
	"regexp"
	"strings"
)

// ValidatePhone 验证手机号合法性
// 规则：
// 1. 不能为空
// 2. 去除空格后长度为 11 位
// 3. 必须以 1 开头
// 4. 第二位为 3-9
// 5. 符合中国大陆手机号格式
func ValidatePhone(phone string) bool {
	// 1. 检查是否为空
	if phone == "" {
		return false
	}

	// 2. 去除空格和横杠
	phone = strings.TrimSpace(phone)
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, " ", "")

	// 3. 检查长度
	if len(phone) != 11 {
		return false
	}

	// 4. 正则表达式验证中国大陆手机号格式
	// 匹配规则：
	// - 以 1 开头
	// - 第二位为 3-9
	// - 后面 9 位为数字
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)

	return phoneRegex.MatchString(phone)
}

// ValidatePassword 验证密码合法性
func ValidatePassword(pwd string) bool {
	length := len(pwd)

	// 长度限制
	if length < 8 || length > 20 {
		return false
	}
	return true

	//var hasLetter, hasNumber bool
	//
	//for _, c := range pwd {
	//	switch {
	//	case c >= 'a' && c <= 'z':
	//		hasLetter = true
	//	case c >= 'A' && c <= 'Z':
	//		hasLetter = true
	//	case c >= '0' && c <= '9':
	//		hasNumber = true
	//	default:
	//		// 不允许特殊字符（如果你想允许，可以删掉这段）
	//		return false
	//	}
	//}
	//
	//// 必须包含字母 + 数字
	//return hasLetter && hasNumber
}
