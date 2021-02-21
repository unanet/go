package oidcprovider

import (
	"context"

	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

type Option func(*Service)

type Service struct {
	Verifier     *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
}

type Config struct {
	OidcConnURL      string `split_words:"true" default:"https://idp.unanet.io/auth/realms/devops"`
	OidcClientID     string `split_words:"true" required:"true"`
	OidcClientSecret string `split_words:"true" required:"true"`
	OidcRedirectURL  string `split_words:"true" required:"true"`
	OidcScopes       string `split_words:"true" default:"profile,email"`
}

func NewService(cfg *Config, opts ...Option) (*Service, error) {

	provider, err := oidc.NewProvider(context.Background(), cfg.OidcConnURL)
	if err != nil {
		return nil, err
	}

	return &Service{
		Verifier: provider.Verifier(&oidc.Config{
			ClientID: cfg.OidcClientID,
		}),
		oauth2Config: oauth2.Config{
			ClientID:     cfg.OidcClientID,
			ClientSecret: cfg.OidcClientSecret,
			RedirectURL:  cfg.OidcRedirectURL,
			Endpoint:     provider.Endpoint(),
			// "openid" is a required scope for OpenID Connect flows.
			Scopes: []string{oidc.ScopeOpenID, cfg.OidcScopes},
		},
	}, nil
}

func (svc *Service) AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string {
	return svc.oauth2Config.AuthCodeURL(state, opts...)
}

func (svc *Service) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	return svc.oauth2Config.Exchange(ctx, code, opts...)
}
