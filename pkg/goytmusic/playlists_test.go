package goytmusic

import (
	"flag"
	"os"
	"testing"
)

// Define a flag -update. Use: go test -update
var update = flag.Bool("update", false, "update json files with API data")

func TestListLikedPlaylistsExtraction(t *testing.T) {
	const filePath = "testdata/liked_playlists.json"

	if *update {
		cookie := os.Getenv("AUTH_COOKIE")
		authUser := os.Getenv("AUTH_USER")
		if cookie == "" {
			t.Fatal("AUTH_COOKIE not configured. Impossible to update testdata.")
		}

		client := NewClient(nil).WithAuth(cookie, authUser)

		req, _ := client.NewRequest("POST", "browse?prettyPrint=false", client.BrowseBody(brIDLikedPlaylists))
		body, _, err := client.Do(req)
		if err != nil {
			t.Fatalf("Error while retrieving liked playlists: %v", err)
		}

		err = os.WriteFile(filePath, body, 0o644)
		if err != nil {
			t.Fatalf("Error while writing golden file: %v", err)
		}
		t.Logf("File %s successfully updated!", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Test file not found. Run with -update: %v", err)
	}

	// TODO: improve verification to find exactly were the change occurred
	playlists := extractPlaylists(data)
	if len(playlists) == 0 {
		t.Error("No playlists found.")
	}
}

func TestGetPlaylistExtraction(t *testing.T) {
	const filePath = "testdata/playlist_tracks.json"
	const playlistBrowseID = "VLPLrG2h7_c1yh1tgTYhJlrXvgEIgGfL96BV" // public playlist

	if *update {
		cookie := os.Getenv("AUTH_COOKIE")
		authUser := os.Getenv("AUTH_USER")
		if cookie == "" {
			t.Fatal("AUTH_COOKIE not configured. Impossible to update testdata.")
		}

		client := NewClient(nil).WithAuth(cookie, authUser)

		req, _ := client.NewRequest("POST", "browse?prettyPrint=false", client.BrowseBody(playlistBrowseID))
		body, _, err := client.Do(req)
		if err != nil {
			t.Fatalf("Error while retrieving playlist: %v", err)
		}

		err = os.WriteFile(filePath, body, 0o644)
		if err != nil {
			t.Fatalf("Error while writing golden file: %v", err)
		}
		t.Logf("File %s successfully updated!", filePath)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Test file not found. Run with -update: %v", err)
	}

	// TODO: improve verification to find exactly were the change occurred
	playlist := extractPlaylistWithTracks(data)
	if len(playlist.Tracks) == 0 {
		t.Error("No track extracted from playlist.")
	}
}
