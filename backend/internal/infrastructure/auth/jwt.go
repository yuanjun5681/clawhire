package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 是 JWT 载荷，只保留登录必要字段。
type Claims struct {
	AccountID   string `json:"sub"`
	AccountType string `json:"type"`
	jwt.RegisteredClaims
}

// JWTIssuer 负责签发与校验 HS256 token。
type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
	issuer string
	now    func() time.Time
}

func NewJWTIssuer(secret string, ttl time.Duration, issuer string) *JWTIssuer {
	return &JWTIssuer{
		secret: []byte(secret),
		ttl:    ttl,
		issuer: issuer,
		now:    time.Now,
	}
}

// SetNow 让测试可注入时钟。
func (j *JWTIssuer) SetNow(now func() time.Time) { j.now = now }

func (j *JWTIssuer) TTL() time.Duration { return j.ttl }

func (j *JWTIssuer) Issue(accountID, accountType string) (string, time.Time, error) {
	now := j.now().UTC()
	expiresAt := now.Add(j.ttl)
	claims := Claims{
		AccountID:   accountID,
		AccountType: accountType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   accountID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt: %w", err)
	}
	return signed, expiresAt, nil
}

func (j *JWTIssuer) Verify(raw string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("token invalid")
	}
	return claims, nil
}
