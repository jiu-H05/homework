// Package httpapi 提供 REST API：路由、中间件与各资源处理器。
package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// 极简 HMAC 签名 Bearer 令牌：base64url(载荷).base64url(签名)
// 载荷内含用户标识、角色与过期时间，服务端无状态校验。

var (
	errInvalidToken = errors.New("令牌无效")
	errTokenExpired = errors.New("令牌已过期")
)

type claims struct {
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RealName string `json:"realName"`
	Exp      int64  `json:"exp"`
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func signToken(c claims, secret []byte) (string, error) {
	body, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	return b64(body) + "." + b64(mac.Sum(nil)), nil
}

func parseToken(token string, secret []byte) (*claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, errInvalidToken
	}
	body, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, errInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, errInvalidToken
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write(body)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return nil, errInvalidToken
	}
	var c claims
	if err := json.Unmarshal(body, &c); err != nil {
		return nil, errInvalidToken
	}
	if time.Now().Unix() > c.Exp {
		return nil, errTokenExpired
	}
	return &c, nil
}
