package auth

import (
	"context"

	"github.com/golang-jwt/jwt"
)

type ctxKeyJwtClaims int

const jwtClaimsID ctxKeyJwtClaims = 0

func Claims(ctx context.Context) jwt.MapClaims {
	if ctx == nil {
		return map[string]interface{}{}
	}

	if m, ok := ctx.Value(jwtClaimsID).(jwt.MapClaims); ok {
		return m
	}

	return map[string]interface{}{}
}

func CtxWithClaims(ctx context.Context, claims jwt.MapClaims) context.Context {
	return context.WithValue(ctx, jwtClaimsID, claims)
}
