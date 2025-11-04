package authenticator

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"log"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	*oidc.Provider
	oauth2.Config
	Audience string
	Domain   string
}

func New(domain string, clientId, clientSecret, audience, redirectUrl string) (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+domain+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email", "offline_access"},
	}

	return &Authenticator{
		Provider: provider,
		Config:   conf,
		Audience: audience,
		Domain:   domain,
	}, nil
}

func (a *Authenticator) VerifyIDToken(ctx context.Context, token *oauth2.Token) (*oidc.IDToken, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, errors.New("no id_token field in oauth2 token")
	}

	oidcConfig := &oidc.Config{
		ClientID: a.ClientID,
	}

	return a.Verifier(oidcConfig).Verify(ctx, rawIDToken)
}

func (a *Authenticator) AuthCodeURLWithPKCE(state string, verifier string) string {
	challenge := generateCodeChallenge(verifier)

	log.Printf("Audience: %v", a.Audience)

	return a.Config.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		oauth2.SetAuthURLParam("audience", a.Audience),
	)
}

func (a *Authenticator) ExchangeWithPKCE(ctx context.Context, code string, codeVerifier string) (*oauth2.Token, error) {
	return a.Config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
}

func (a *Authenticator) RefreshAccessToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
	}
	ts := a.Config.TokenSource(ctx, token)

	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}

	return newToken, nil
}

func (a *Authenticator) IsExpired(accessToken string) bool {
	token, _, err := new(jwt.Parser).ParseUnverified(accessToken, jwt.MapClaims{})
	if err != nil {
		return true
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return true
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return true
	}

	return time.Unix(int64(exp), 0).Before(time.Now())
}

func (a *Authenticator) IsValidAccessToken(ctx context.Context, tokenStr string) (bool, error) {
	provider, err := oidc.NewProvider(ctx, "https://"+a.Domain+"/")
	if err != nil {
		return false, err
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: a.Audience,
	})

	_, err = verifier.Verify(ctx, tokenStr)
	if err != nil {
		return false, err
	}

	return true, nil
}

func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}
