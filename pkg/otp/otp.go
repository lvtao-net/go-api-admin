package otp

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync"
	"time"
)

// OTP 存储
type OTPStore struct {
	Code      string
	Email     string
	Type      string // register, password-reset, email-login
	ExpiresAt time.Time
}

var (
	store sync.Map
	// 默认验证码有效期 10 分钟
	defaultExpiry = 10 * time.Minute
	// 验证码长度
	codeLength = 6
)

// Generate 生成验证码
func Generate(email, otpType string) string {
	code := generateCode(codeLength)
	key := getKey(email, otpType)

	store.Store(key, OTPStore{
		Code:      code,
		Email:     email,
		Type:      otpType,
		ExpiresAt: time.Now().Add(defaultExpiry),
	})

	return code
}

// Verify 验证验证码
func Verify(email, otpType, code string) (bool, error) {
	key := getKey(email, otpType)

	value, ok := store.Load(key)
	if !ok {
		return false, errors.New("验证码不存在或已过期")
	}

	otp, ok := value.(OTPStore)
	if !ok {
		return false, errors.New("验证码格式错误")
	}

	// 检查是否过期
	if time.Now().After(otp.ExpiresAt) {
		store.Delete(key)
		return false, errors.New("验证码已过期")
	}

	// 验证码匹配
	if otp.Code != code {
		return false, errors.New("验证码错误")
	}

	// 验证成功后删除
	store.Delete(key)
	return true, nil
}

// Delete 删除验证码
func Delete(email, otpType string) {
	key := getKey(email, otpType)
	store.Delete(key)
}

// getKey 生成存储键
func getKey(email, otpType string) string {
	return otpType + ":" + email
}

// generateCode 生成数字验证码
func generateCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

// SetExpiry 设置验证码有效期（用于测试）
func SetExpiry(d time.Duration) {
	defaultExpiry = d
}
