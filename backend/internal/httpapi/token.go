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
// 载荷内含用户标识、角色、令牌类型与过期时间，服务端无状态校验。

var (
	errInvalidToken = errors.New("令牌无效")
	errTokenExpired = errors.New("令牌已过期")
)

type claims struct {
	UID      int64  `json:"uid"`
	Username string `json:"username"`
	Role     string `json:"role"`
	RealName string `json:"realName"`
	Type     string `json:"typ"` // access | refresh
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

// issueTokens 同时签发短期 access 与长期 refresh 令牌。
func issueTokens(u claims, secret []byte) (access, refresh string, err error) {
	access, err = signToken(claims{
		UID: u.UID, Username: u.Username, Role: u.Role, RealName: u.RealName,
		Type: "access", Exp: time.Now().Add(2 * time.Hour).Unix(),
	}, secret)
	if err != nil {
		return "", "", err
	}
	refresh, err = signToken(claims{
		UID: u.UID, Username: u.Username, Role: u.Role, RealName: u.RealName,
		Type: "refresh", Exp: time.Now().Add(7 * 24 * time.Hour).Unix(),
	}, secret)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
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
