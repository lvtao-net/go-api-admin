package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lvtao/go-gin-api-admin/internal/config"
)

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AdminClaims struct {
	AdminID string `json:"admin_id"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken 生成用户Token
func GenerateToken(userID, email, role string) (string, error) {
	cfg := config.GetConfig()
	
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.Expires) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-gin-api-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GenerateAdminToken 生成管理员Token
func GenerateAdminToken(adminID, email string) (string, error) {
	cfg := config.GetConfig()
	
	claims := AdminClaims{
		AdminID: adminID,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.Expires) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-gin-api-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ValidateToken 验证用户Token
func ValidateToken(tokenString string) (*Claims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateUserToken 验证Auth集合用户Token
func ValidateUserToken(tokenString string) (*UserClaims, error) {
	cfg := config.GetConfig()

	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateAdminToken 验证管理员Token
func ValidateAdminToken(tokenString string) (*AdminClaims, error) {
	cfg := config.GetConfig()
	
	token, err := jwt.ParseWithClaims(tokenString, &AdminClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*AdminClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// UserClaims 用户Token声明
type UserClaims struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	Collection string `json:"collection"`
	jwt.RegisteredClaims
}

// GenerateRefreshToken 生成刷新Token
func GenerateRefreshToken(userID string) (string, error) {
	cfg := config.GetConfig()
	
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.RefreshExpires) * time.Second)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "go-gin-api-admin-refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GenerateUserToken 生成用户Token（Auth集合）
func GenerateUserToken(userID, email, collection string) (string, error) {
	cfg := config.GetConfig()

	claims := UserClaims{
		UserID:     userID,
		Email:      email,
		Collection: collection,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.Expires) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-gin-api-admin",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// GenerateRefreshTokenWithInfo 生成带用户信息的刷新Token
func GenerateRefreshTokenWithInfo(email, collection string) (string, error) {
	cfg := config.GetConfig()
	
	claims := UserClaims{
		Email:      email,
		Collection: collection,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JWT.RefreshExpires) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-gin-api-admin-refresh",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWT.Secret))
}

// ValidateRefreshToken 验证刷新Token
func ValidateRefreshToken(tokenString string) (map[string]interface{}, error) {
	cfg := config.GetConfig()
	
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.JWT.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		return map[string]interface{}{
			"email":      claims.Email,
			"collection": claims.Collection,
		}, nil
	}

	return nil, errors.New("invalid refresh token")
}
