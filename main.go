package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

// Global constants
const TOKEN_FILENAME = "token.json"

func main() {
	ctx := context.Background()

	service, err := getService(ctx)
	if err != nil {
		panic(err)
	}

	call := service.Playlists.List([]string{"snippet"})
	call.Mine(true)
	resp, err := call.Do()
	if err != nil {
		panic(err)
	}
	_ = resp
}

func getService(ctx context.Context) (*youtube.Service, error) {
	// Read client secret from json
	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		return nil, err
	}

	// Get OAuth2 config
	conf, err := google.ConfigFromJSON(b, youtube.YoutubeScope)
	if err != nil {
		return nil, err
	}

	// Get token
	tok, err := getToken(conf, ctx)
	if err != nil {
		return nil, err
	}

	return youtube.NewService(ctx, option.WithTokenSource(conf.TokenSource(ctx, tok)))
}

func getToken(conf *oauth2.Config, ctx context.Context) (*oauth2.Token, error) {
	// Check if we have a locally cached token
	b, err := os.ReadFile(TOKEN_FILENAME)
	if err == nil {
		// Unmarshal json and return
		tok := oauth2.Token{}
		err := json.Unmarshal(b, &tok)
		if err == nil {
			return &tok, nil
		}
		couldnt("unmarshal token json", err)
	} else {
		couldnt("read token file", err)
	}

	// Need to get a fresh token
	return getNewToken(conf, ctx)
}

func getNewToken(conf *oauth2.Config, ctx context.Context) (*oauth2.Token, error) {
	conf.RedirectURL = "http://localhost:8080"

	// Start HTTP server to receive auth code
	codeChan := make(chan string)
	go http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		codeChan <- r.URL.Query().Get("code")
		w.Write([]byte("You may now close this tab/window. :)"))
	}))

	// Redirect user to consent page
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Please open in a browser: %v\n", url)

	// Get code from HTTP server
	code := <-codeChan

	// Get token
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	// Cache token for later
	b, err := json.Marshal(tok)
	if err == nil {
		err = os.WriteFile(TOKEN_FILENAME, b, os.ModePerm)
		if err != nil {
			couldnt("write token to file", err)
		}
	} else {
		couldnt("marshal token to json", err)
	}

	return tok, nil
}

func couldnt(action string, err error) {
	fmt.Printf("WARNING: couldn't %s: %s\n", action, err.Error())
}
