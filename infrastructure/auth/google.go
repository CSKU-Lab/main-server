package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CSKU-Lab/main-server/configs"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type googleOauthHandler struct {
	auth   *oauth2.Config
	config *configs.Config
	states map[string]bool
}

func NewGoogleAuth(c *configs.Config) *googleOauthHandler {
	auth := &oauth2.Config{
		ClientID:     c.GoogleClientID,
		ClientSecret: c.GoogleClientSecret,
		RedirectURL:  fmt.Sprintf("%v/api/v1/auth/sign-in/google/callback", c.ApiURL),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &googleOauthHandler{auth: auth, states: make(map[string]bool), config: c}
}

func (g *googleOauthHandler) GenerateAuthURL() (string, error) {
	state, err := generateState()
	if err != nil {
		return "", err
	}

	g.states[state] = true
	return g.auth.AuthCodeURL(state), nil
}

func (g *googleOauthHandler) VerifyState(state string) bool {
	if g.states[state] {
		delete(g.states, state)
		return true
	}
	return false
}

var userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo"

func (g *googleOauthHandler) GetUserInfo(ctx context.Context, code string) (*UserInfo, error) {
	token, err := g.auth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client := g.auth.Client(ctx, token)
	resp, err := client.Get(userInfoURL)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	var userInfo map[string]interface{}

	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		return nil, err
	}

	return &UserInfo{
		Email:        userInfo["email"].(string),
		ProfileImage: userInfo["picture"].(string),
	}, nil
}
