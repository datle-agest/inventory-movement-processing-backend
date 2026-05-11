package jwtc

import (
	"context"
	"errors"
	"flag"
	"fmt"
	sctx "inventory-movement-processing/pkg/service_context"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultSecret               = "very-important-please-change-it!" // in 32 bytes
	defaultExpireTokenInSeconds = 60 * 60 * 24 * 7                   // 7d
)

var (
	ErrSecretKeyNotValid     = errors.New("secret key must be in 32 bytes")
	ErrTokenLifeTimeTooShort = errors.New("token life time too short")
)

type jwtx struct {
	id                   string
	secret               string
	expireTokenInSeconds int
}

func NewJWT(id string) *jwtx {
	return &jwtx{
		id: id,
	}
}

func (j *jwtx) ID() string {
	return j.id
}

func (j *jwtx) InitFlags() {
	flag.StringVar(
		&j.secret,
		"jwt-secret",
		defaultSecret,
		"Secret key to sign JWT")
	flag.IntVar(
		&j.expireTokenInSeconds,
		"jwt-exp-secs",
		defaultExpireTokenInSeconds,
		"Token life time in second")
}

func (j *jwtx) Activate(_ sctx.ServiceContext) error {
	if len(j.secret) < 32 {
		return ErrSecretKeyNotValid
	}
	if j.expireTokenInSeconds < 60 {
		return ErrTokenLifeTimeTooShort
	}
	return nil
}

func (j *jwtx) Stop() error {
	return nil
}

func (j *jwtx) IssueToken(ctx context.Context, id, sub string, seconds int) (token string, expSecs int, err error) {
	now := time.Now().UTC()

	exp := j.expireTokenInSeconds
	if seconds > 0 {
		exp = seconds
	}

	claims := jwt.RegisteredClaims{
		Subject:   sub,
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Second * time.Duration(exp))),
		NotBefore: jwt.NewNumericDate(now),
		IssuedAt:  jwt.NewNumericDate(now),
		ID:        id,
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSignedStr, err := t.SignedString([]byte(j.secret))

	if err != nil {
		return "", 0, err
	}

	return tokenSignedStr, exp, nil
}

func (j *jwtx) ParseToken(ctx context.Context, tokenString string) (*jwt.RegisteredClaims, error) {
	if j == nil {
		return nil, errors.New("jwt component is nil")
	}

	var rc jwt.RegisteredClaims
	token, err := jwt.ParseWithClaims(tokenString, &rc, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if token == nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return &rc, nil
}
