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
	"golang.org/x/oauth2"
)

type Option func(*Service)

type Service struct {
	// TODO: Remove after apps are migrated from legacy /login route
	jwtAuth      *jwtauth.JWTAuth // Used for Legacy login route
	verifier     *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
}

type Config struct {
	ConnURL      string `split_words:"true" required:"true"`
	ClientID     string `split_words:"true" required:"true"`
	ClientSecret string `split_words:"true" required:"true"`
	RedirectURL  string `split_words:"true" required:"true"`
}

func JWTClientOpt(signingKey string) Option {
	return func(svc *Service) {
		svc.jwtAuth = jwtauth.New(jwt.SigningMethodHS256.Alg(), []byte(signingKey), nil)
	}
}

func WithAdditionalScopesOpt(scopes []string) Option {
	return func(svc *Service) {
		svc.oauth2Config.Scopes = append(svc.oauth2Config.Scopes, scopes...)
	}
}

func WithScopesOpt(scopes []string) Option {
	return func(svc *Service) {
		svc.oauth2Config.Scopes = scopes
	}
}

func NewService(cfg Config, opts ...Option) (*Service, error) {

	provider, err := oidc.NewProvider(context.Background(), cfg.ConnURL)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		verifier: provider.Verifier(&oidc.Config{
			ClientID: cfg.ClientID,
		}),
		oauth2Config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     provider.Endpoint(),
			// "openid" is a required scope for OpenID Connect flows.
			Scopes: []string{oidc.ScopeOpenID, "profile", "email"},
		},
	}

	for _, opt := range opts {
		opt(svc)
	}

	return svc, nil
}

// TokenVerification verifies the incoming token request
// right now we are supporting the admin token, the k8s /login token from cloud-api and the new keycloak token
// TODO: remove support for k8s cloud-api /login and the admin token
func (svc *Service) TokenVerification(r *http.Request) (jwt.MapClaims, error) {
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

func (svc *Service) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return svc.oauth2Config.AuthCodeURL(state, opts...)
}

func (svc *Service) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return svc.oauth2Config.Exchange(ctx, code, opts...)
}

func (svc *Service) Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return svc.verifier.Verify(ctx, rawIDToken)
}
