package middleware

import (
	"context"
	"fmt"
	"net/http"

	"github.com/unanet/go/v2/pkg/auth"

	"github.com/casbin/casbin/v2"
	"github.com/go-chi/jwtauth/v5"
	"github.com/go-chi/render"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/unanet/go/v2/pkg/errors"
	"github.com/unanet/go/v2/pkg/identity"
)

func extractRoles(ctx context.Context, claims jwt.MapClaims) []interface{} {
	Log(ctx).Debug("extract role from incoming claims", zap.Any("claims", claims))
	if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := ra["roles"].([]interface{}); ok {
			Log(ctx).Debug("incoming claim roles slice found", zap.Any("role", roles))
			return roles
		}
	}

	Log(ctx).Debug("unknown role extracted")
	return []interface{}{}
}

func AuthenticationMiddleware(adminToken string, idv *identity.Validator, enforcer *casbin.Enforcer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Admin token, you shall PASS!!!
			if jwtauth.TokenFromHeader(r) == adminToken {
				ctx = auth.CtxWithClaims(ctx, map[string]interface{}{
					"sub": "admin",
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			claims, err := idv.Validate(r)
			if err != nil {
				Log(ctx).Debug("failed token verification", zap.Error(err))
				render.Respond(w, r, err)
				return
			}

			Log(ctx).Debug("incoming auth claims", zap.Any("claims", claims))

			var grantedAccess bool

			Log(ctx).Debug(fmt.Sprintf("checking auth for URL = %s, Method = %s", r.URL.Path, r.Method))

			// Range over the roles to see if we have access to the resource
			for _, role := range extractRoles(ctx, claims) {
				grantedAccess, err = enforcer.Enforce(role, r.URL.Path, r.Method)
				if err != nil {
					Log(ctx).Error("casbin enforced resulted in an error", zap.Error(err))
					render.Status(r, 500)
					return
				}

				if grantedAccess {
					Log(ctx).Debug(fmt.Sprintf("access granted. Role = %s, URL = %s, Method = %s", role, r.URL.Path, r.Method))
					break
				}
			}

			if !grantedAccess {
				Log(ctx).Debug(fmt.Sprintf("not authorized. URL = %s, Method = %s", r.URL.Path, r.Method))
				render.Respond(w, r, errors.NewRestError(403, "Forbidden"))
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.CtxWithClaims(ctx, claims)))
		}
		return http.HandlerFunc(hfn)
	}
}
