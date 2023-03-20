package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
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

var PLAYLIST_ITEMS_PARTS = []string{"snippet"}

func main() {
	err := main2()
	if err != nil {
		// Force new token and retry
		os.Remove(TOKEN_FILENAME)
		err = main2()
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main2() error {
	ctx := context.Background()
	s, err := getService(ctx)
	if err != nil {
		return err
	}

	playlists, err := getPlaylists(s)
	if err != nil {
		return err
	}

	playlistID, err := promptForPlaylist(playlists)
	if err != nil {
		return err
	}

	// items, err := getPlaylistItems(s, playlistID)
	// 	if err != nil {
	// 	return err
	// }

	return sort(s, playlistID)
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

func getPlaylists(s *youtube.Service) (*youtube.PlaylistListResponse, error) {
	listPlaylistsCall := s.Playlists.List([]string{"snippet", "contentDetails"})
	listPlaylistsCall.Mine(true)
	return listPlaylistsCall.Do()
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

func getPlaylistItems(s *youtube.Service, playlistID string) ([]*youtube.PlaylistItem, error) {
	items := []*youtube.PlaylistItem{}
	nextPageToken := ""

	for {
		// Call API
		listItemsCall := s.PlaylistItems.List(PLAYLIST_ITEMS_PARTS)
		listItemsCall.PlaylistId(playlistID)
		listItemsCall.MaxResults(50)
		listItemsCall.PageToken(nextPageToken)

		ret, err := listItemsCall.Do()
		if err != nil {
			return nil, err
		}

		// Append items to array
		items = append(items, ret.Items...)

		// Check if there's a next page
		if ret.NextPageToken == "" {
			break
		} else {
			nextPageToken = ret.NextPageToken
		}
	}

	return items, nil
}

func updatePlaylistItems(s *youtube.Service, item *youtube.PlaylistItem) (*youtube.PlaylistItem, error) {
	updateItemsCall := s.PlaylistItems.Update(PLAYLIST_ITEMS_PARTS, item)
	return updateItemsCall.Do()
}

// Sorting function

// Use insertion sort for simplicity
func sort(s *youtube.Service, playlistID string) error {
	next := 1 // next index to sort

	for {
		// TODO: don't need to continually get, just update a local copy
		items, err := getPlaylistItems(s, playlistID)
		if err != nil {
			return err
		}

		if next >= len(items) {
			break
		}

		// Find where items[next] should be inserted
		for i := 0; i < next; i++ {
			if items[i].Snippet.Title > items[next].Snippet.Title {
				// Update snippet.position
				items[next].Snippet.Position = int64(i)
				fmt.Printf("moving %q from pos %d to %d\n", items[next].Snippet.Title, next, i)
				_, err = updatePlaylistItems(s, items[next])
				if err != nil {
					return err
				}
				break
			}
		}
		// If we escaped the loop without updating anything, then items[next] is
		// already in the right place. So nothing to do here except go on to
		// the next iteration.
		next++
	}
	return nil
}

// Helper functions

func dump(v any, filename string) {
	os.WriteFile(filename, []byte(pretty.Sprint(v)), os.ModePerm)
}

func couldnt(action string, err error) {
	fmt.Printf("WARNING: couldn't %s: %s\n", action, err.Error())
}
