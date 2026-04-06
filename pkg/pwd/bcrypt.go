package pwd

import (
	"golang.org/x/crypto/bcrypt"
)

// GenerateFromPassword 加密密码
func GenerateFromPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword), err
}

// CompareHashAndPassword 比较密码
func CompareHashAndPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
}
