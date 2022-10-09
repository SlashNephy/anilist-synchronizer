package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/goccy/go-json"
	"github.com/labstack/gommon/random"
	"golang.org/x/oauth2"
)

func getOAuth2Config(config *config) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     config.AniListClientID,
		ClientSecret: config.AniListClientSecret,
		RedirectURL:  "https://anilist.co/api/v2/oauth/pin",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://anilist.co/api/v2/oauth/authorize",
			TokenURL: "https://anilist.co/api/v2/oauth/token",
		},
	}
}

func acquireToken(ctx context.Context, oauth *oauth2.Config) (*oauth2.Token, error) {
	state := random.String(64)
	url := oauth.AuthCodeURL(state)
	fmt.Printf("Open URL & Paste code:\n=> %s\n", url)

	var code string
	_, err := fmt.Scanln(&code)
	if err != nil {
		return nil, err
	}

	token, err := oauth.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func storeToken(token *oauth2.Token, filename string) error {
	tokenJson, err := json.Marshal(token)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filename, tokenJson, 0666); err != nil {
		return err
	}

	return nil
}

func loadToken(filename string) (*oauth2.Token, error) {
	tokenJson, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var token *oauth2.Token
	if err = json.Unmarshal(tokenJson, &token); err != nil {
		return nil, err
	}

	return token, nil
}

func getOAuth2Client(ctx context.Context, config *config, tokenFile string) (*http.Client, error) {
	oauth := getOAuth2Config(config)

	var token *oauth2.Token
	if _, err := os.Stat(tokenFile); errors.Is(err, os.ErrNotExist) {
		token, err = acquireToken(ctx, oauth)
		if err != nil {
			return nil, err
		}

		if err = storeToken(token, tokenFile); err != nil {
			return nil, err
		}
	} else {
		token, err = loadToken(tokenFile)
		if err != nil {
			return nil, err
		}
	}

	source := oauth.TokenSource(ctx, token)
	return oauth2.NewClient(ctx, source), nil
}
