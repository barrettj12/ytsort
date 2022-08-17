package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gosuri/uitable"
	"github.com/kr/pretty"
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
	panicIfNotNil(err)

	listPlaylistsCall := service.Playlists.List([]string{"snippet", "contentDetails"})
	listPlaylistsCall.Mine(true)
	playlists, err := listPlaylistsCall.Do()
	panicIfNotNil(err)

	playlistID, err := promptForPlaylist(playlists)
	panicIfNotNil(err)

	listItemsCall := service.PlaylistItems.List([]string{"contentDetails", "id", "snippet", "status"})
	listItemsCall.PlaylistId(playlistID)
	listItemsCall.MaxResults(50)
	items, err := listItemsCall.Do()
	panicIfNotNil(err)

	dump(items, "items.txt")

	// If `items.NextPageToken != ""`: then we need to get next page

	fmt.Println("This playlist contains the following items:")
	for _, it := range items.Items {
		fmt.Printf("  - %s\n", it.Snippet.Title)
	}
}

func panicIfNotNil(v any) {
	if v != nil {
		panic(v)
	}
}

func dump(v any, filename string) {
	os.WriteFile(filename, []byte(pretty.Sprint(v)), os.ModePerm)
}

// returns playlist ID
func promptForPlaylist(r *youtube.PlaylistListResponse) (string, error) {
	fmt.Println("Which of the following playlists would you like to sort?")

	table := uitable.New()
	table.MaxColWidth = 35
	table.AddRow("#", "Name", "Len", "ID")

	for i, pl := range r.Items {
		table.AddRow(
			i,
			pl.Snippet.Title,
			pl.ContentDetails.ItemCount,
			pl.Id,
		)
	}
	fmt.Println(table)

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("Enter number (#): ")
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		return "", err
	}

	n, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return "", err
	}

	return r.Items[n].Id, nil
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
