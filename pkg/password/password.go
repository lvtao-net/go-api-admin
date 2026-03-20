package password

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword 对密码进行 bcrypt 加密
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword 验证密码是否匹配
// hashedPassword 是加密后的密码
// password 是用户输入的明文密码
func VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// MustHashPassword 对密码进行加密，如果失败则 panic
func MustHashPassword(password string) string {
	hash, err := HashPassword(password)
	if err != nil {
		panic(err)
	}
	return hash
}
