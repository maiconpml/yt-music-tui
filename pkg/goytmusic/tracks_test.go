package goytmusic

import (
	"os"
	"testing"
)

// Define a flag -update. Use: go test -update

func TestExtractTracksFromQueue(t *testing.T) {
	const filePath = "testdata/next_tracks_playlist.json"

	if *update {
		cookie := os.Getenv("AUTH_COOKIE")
		authUser := os.Getenv("AUTH_USER")
		if cookie == "" {
			t.Fatal("AUTH_COOKIE not configured. Impossible to update testdata.")
		}

		client := NewClient(nil).WithAuth(cookie, authUser)

		body := struct {
			PlaylistID string  `json:"playlistId"`
			Params     string  `json:"params"`
			Context    Context `json:"context"`
		}{"PLrG2h7_c1yh1Vr8hVtm5Xsp10gtphPI5S", "", client.commonContext}
		req, _ := client.NewRequest("POST", "next?prettyPrint=false", body)
		respBody, _, err := client.Do(req)
		if err != nil {
			t.Fatalf("Error while retrieving liked playlists: %v", err)
		}

		err = os.WriteFile(filePath, respBody, 0o644)
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
	tracks, _ := extractTracksFromQueue(data)
	if len(tracks) == 0 {
		t.Error("No tracks found.")
	}
}
