package identity

import (
	"context"
	goErrors "errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc"
	"github.com/go-chi/jwtauth"
	"github.com/golang-jwt/jwt"

	"github.com/unanet/go/pkg/errors"
)

type ValidatorOption func(validator *Validator)

type ValidatorConfig struct {
	ClientID      string `split_words:"true" required:"true"`
	ConnectionURL string `split_words:"true" required:"true"`
}

type Validator struct {
	// TODO: Remove after apps are migrated from legacy /login route
	jwtAuth  *jwtauth.JWTAuth // Used for Legacy login route
	verifier *oidc.IDTokenVerifier
}

func JWTClientValidatorOpt(signingKey string) ValidatorOption {
	return func(v *Validator) {
		v.jwtAuth = jwtauth.New(jwt.SigningMethodHS256.Alg(), []byte(signingKey), nil)
	}
}

func NewValidator(cfg ValidatorConfig, opts ...ValidatorOption) (*Validator, error) {
	provider, err := oidc.NewProvider(context.Background(), cfg.ConnectionURL)
	if err != nil {
		return nil, err
	}

	validator := Validator{
		verifier: provider.Verifier(&oidc.Config{
			ClientID: cfg.ClientID,
		}),
	}

	for _, opt := range opts {
		opt(&validator)
	}

	return &validator, nil
}

// TokenVerification verifies the incoming token request
// right now we are supporting the admin token, the k8s /login token from cloud-api and the new keycloak token
// TODO: remove support for k8s cloud-api /login and the admin token
func (svc *Validator) Validate(r *http.Request) (jwt.MapClaims, error) {
	ctx := r.Context()
	token := jwtauth.TokenFromHeader(r)
	// Empty Token return unauthorized error
	if len(token) == 0 {
		return nil, errors.ErrEmptyToken
	}

	// Attempt to verify the token again OIDC provider (Keycloak via Okta auth) first
	// If its a valid token (no error) return immediately
	keyCloakToken, verr := svc.verifier.Verify(ctx, token)
	if verr != nil {
		if goErrors.Is(verr, jwtauth.ErrExpired) {
			return nil, errors.ErrExpired
		}
	} else {
		var idTokenClaims = new(jwt.MapClaims)
		if err := keyCloakToken.Claims(&idTokenClaims); err != nil {
			return nil, errors.ErrMapTokenClaims
		}
		return *idTokenClaims, nil
	}

	// Attempt to verify the token against cloud-api /login route for k8s tokens
	// TODO: eventually remove this when everything is using Keycloak instead of /login route in cloud-api
	if svc.jwtAuth != nil {
		ct, err := jwtauth.VerifyRequest(svc.jwtAuth, r, jwtauth.TokenFromHeader)
		if err != nil {
			if goErrors.Is(err, jwtauth.ErrExpired) {
				return nil, errors.ErrExpired
			}
			return nil, errors.NewRestError(http.StatusUnauthorized, fmt.Sprintf("Unauthorized: %s", err.Error()))
		}
		if ct == nil || !ct.Valid {
			return nil, errors.ErrInvalidToken
		}
		if tokenClaims, ok := ct.Claims.(jwt.MapClaims); ok {
			return tokenClaims, nil
		} else {
			return nil, errors.ErrMapTokenClaims
		}
	}
	return nil, errors.ErrUnauthorized
}
