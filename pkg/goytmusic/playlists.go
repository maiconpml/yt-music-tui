package goytmusic

import (
	"errors"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	brIDLikedPlaylists = "FEmusic_liked_playlists"
)

type PlaylistsService service

type Playlist struct {
	Name    string
	ID      string
	NTracks int
	Tracks  []*Track
	Author  *User
}

// ListLiked retrieves and returns an array of Playlist. This array
// corresponds the current user's list of liked playlists.
func (s *PlaylistsService) ListLiked() ([]*Playlist, error) {
	if s.client.isGuest {
		return nil, errors.New("Client is not authenticated")
	}
	u := "browse"
	body := s.client.BrowseBody(brIDLikedPlaylists)
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, err
	}

	respBody, _, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	items := extractPlaylists(respBody)

	return items[1:], nil
}

// Get retrieves and returns the Playlist having the provided id.
func (s *PlaylistsService) Get(id *string) (*Playlist, error) {
	if s.client.isGuest {
		return nil, errors.New("Client is not authenticated")
	}
	u := "browse"
	body := s.client.BrowseBody("VL" + *id)
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, err
	}

	respBody, _, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	pl := extractPlaylistWithTracks(respBody)
	pl.ID = *id

	return pl, nil
}

// Parses a JSON in []byte format into an array of Playlist pointers
// Expects the brIDLikedPlaylists endpoint JSON reponse
func extractPlaylists(b []byte) []*Playlist {
	results := gjson.GetBytes(b, joinPaths(pSingleColumn, pTab0, pTabRendererContent, pSectionList, pContent0, pGridRendererItems))

	var playlists []*Playlist
	results.ForEach(func(key, value gjson.Result) bool {
		if pl := extractPlaylist(value); pl != nil {
			playlists = append(playlists, pl)
		}
		return true
	})

	return playlists
}

// Parses res into a Playlist without loading the tracks
// Expects the playlist contained in the brIDLikedPlaylists
// endpoint JSON reponse
func extractPlaylist(res gjson.Result) *Playlist {
	render := res.Get(pMusicTwoRow)
	if !render.Exists() {
		return nil
	}

	pl := &Playlist{
		Name: render.Get(joinPaths(pTitle, pRun0, pText)).String(),
		ID:   strings.TrimPrefix(render.Get(joinPaths(pTitle, pRun0, pNavEndpoint, pBrowseEnd, pBrowseID)).String(), "VL"),
	}

	author := render.Get(joinPaths(pSubtitle, pRun0))
	if author.Exists() {
		pl.Author = extractUser(author)
	}

	nTracksStr := strings.Fields(render.Get(joinPaths(pSubtitle, pRun2, pText)).String())

	pl.NTracks = 0
	if len(nTracksStr) > 0 {
		pl.NTracks, _ = strconv.Atoi(nTracksStr[0])
	}

	return pl
}

// Parses the JSON in b into a Playlist
// Expects the playlist of the browseId=VLPL... endpoint JSON response
func extractPlaylistWithTracks(b []byte) *Playlist {
	tracks := gjson.GetBytes(b, joinPaths(pTwoColumn, pSecContents, pSectionList, pContent0, pPlaylistShelf, pContents))
	plHeader := gjson.GetBytes(b, joinPaths(pTwoColumn, pTab0, pTabRendererContent, pSectionList, pContent0))
	plHeaderAux := plHeader.Get(pMusicEditablePlaylistHeader)
	if plHeaderAux.Exists() {
		plHeader = plHeaderAux
	}
	plHeader = plHeader.Get(pMusicResponsiveHeader)

	pl := &Playlist{}

	pl.Name = plHeader.Get(joinPaths(pTitle, pRun0, pText)).String()
	pl.Author = extractUser(plHeader)
	tracks.ForEach(func(key, value gjson.Result) bool {
		if tr := extractTrack(value); tr != nil {
			pl.Tracks = append(pl.Tracks, tr)
		}
		return true
	})
	return pl
}
