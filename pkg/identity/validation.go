package identity

import (
	"context"
	goErrors "errors"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/jwtauth/v5"
	"github.com/golang-jwt/jwt/v5"

	"github.com/unanet/go/v2/pkg/errors"
)

type ValidatorOption func(validator *Validator)

type ValidatorConfig struct {
	ClientID      string `split_words:"true" required:"true"`
	ConnectionURL string `split_words:"true" required:"true"`
	Issuer        string `split_words:"true" required:"false"`

	// Optional Skip Checks
	SkipClientIDCheck bool `split_words:"true" default:"false"`
	SkipExpiryCheck   bool `split_words:"true" default:"false"`
	SkipIssuerCheck   bool `split_words:"true" default:"false"`
}

type Validator struct {
	jwtAuth  *jwtauth.JWTAuth
	verifier *oidc.IDTokenVerifier
}

func JWTClientValidatorOpt(signingKey string) ValidatorOption {
	return func(v *Validator) {
		v.jwtAuth = jwtauth.New(jwt.SigningMethodHS256.Alg(), []byte(signingKey), nil)
	}
}

func NewValidator(cfg ValidatorConfig, opts ...ValidatorOption) (*Validator, error) {
	ctx := context.Background()
	if cfg.Issuer != "" {
		// this allows you to manually set the issuer when the connection url is
		// an internal url (needed for hitting the endpoint inside the same cluster in k8s
		ctx = oidc.InsecureIssuerURLContext(ctx, cfg.Issuer)
	}

	provider, err := oidc.NewProvider(ctx, cfg.ConnectionURL)
	if err != nil {
		return nil, err
	}

	validator := Validator{
		verifier: provider.Verifier(&oidc.Config{
			ClientID:          cfg.ClientID,
			SkipClientIDCheck: cfg.SkipClientIDCheck,
			SkipExpiryCheck:   cfg.SkipExpiryCheck,
			SkipIssuerCheck:   cfg.SkipIssuerCheck,
		}),
	}

	for _, opt := range opts {
		opt(&validator)
	}

	return &validator, nil
}

// Validate verifies the incoming token request
func (svc *Validator) Validate(r *http.Request) (jwt.MapClaims, error) {
	ctx := r.Context()
	token := jwtauth.TokenFromHeader(r)
	// Empty Token return unauthorized error
	if len(token) == 0 {
		return nil, errors.ErrEmptyToken
	}

	// Attempt to verify the token again OIDC provider (Keycloak via Okta auth) first
	// If it's a valid token (no error) return immediately
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

	if svc.jwtAuth != nil {
		ct, err := jwtauth.VerifyRequest(svc.jwtAuth, r, jwtauth.TokenFromHeader)
		if err != nil {
			if goErrors.Is(err, jwtauth.ErrExpired) {
				return nil, errors.ErrExpired
			}
			return nil, errors.NewRestError(http.StatusUnauthorized, fmt.Sprintf("Unauthorized: %s", err.Error()))
		}

		if claims, err := ct.AsMap(ctx); err == nil {
			return claims, nil
		} else {
			return nil, errors.ErrMapTokenClaims
		}
	}

	return nil, errors.ErrUnauthorized
}
