package goytmusic

import (
	"os"
	"testing"
)

// Define a flag -update. Use: go test -update

func TestListLikedAlbumsExtraction(t *testing.T) {
	const filePath = "testdata/liked_albums.json"

	if *update {
		cookie := os.Getenv("AUTH_COOKIE")
		if cookie == "" {
			t.Fatal("AUTH_COOKIE not configured. Impossible to update testdata.")
		}

		client := NewClient(nil).WithAuthCookie(cookie)

		req, _ := client.NewRequest("POST", "browse?prettyPrint=false", client.BrowseBody(brIDLikedAlbums))
		body, _, err := client.Do(req)
		if err != nil {
			t.Fatalf("Error while retrieving liked albums: %v", err)
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
	playlists := extractAlbumsFromLibrary(data)
	if len(playlists) == 0 {
		t.Error("No albums found.")
	}
}

func TestGetAlbumExtraction(t *testing.T) {
	const filePath = "testdata/album_tracks.json"
	const playlistBrowseID = "MPREb_2az20wG4HxN" // public playlist

	if *update {
		cookie := os.Getenv("AUTH_COOKIE")
		if cookie == "" {
			t.Fatal("AUTH_COOKIE not configured. Impossible to update testdata.")
		}

		client := NewClient(nil).WithAuthCookie(cookie)

		req, _ := client.NewRequest("POST", "browse?prettyPrint=false", client.BrowseBody(playlistBrowseID))
		body, _, err := client.Do(req)
		if err != nil {
			t.Fatalf("Error while retrieving album: %v", err)
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
	playlist := extractAlbumWithTracks(data)
	if len(playlist.Tracks) == 0 {
		t.Error("No track extracted from album.")
	}
}
