package goytmusic

import (
	"errors"
	"strconv"
	"strings"

	"github.com/tidwall/gjson"
)

const (
	brIDLikedAlbums = "FEmusic_liked_albums"
)

type AlbumService service

type Album struct {
	ID     string
	Name   string
	Year   int
	Tracks []*Track
	Author *User
}

// ListLiked retrieves and returns an array of Playlist. This array
// corresponds the current user's list of liked playlists.
func (s *AlbumService) ListLiked() ([]*Album, error) {
	if s.client.isGuest {
		return nil, errors.New("Client is not authenticated")
	}
	u := "browse"
	body := s.client.BrowseBody(brIDLikedAlbums)
	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, err
	}

	respBody, _, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}

	items := extractAlbumsFromLibrary(respBody)

	return items, nil
}

// Parses a JSON in []byte format into an array of Album pointers
// Expects the brIDLikedAlbums endpoint JSON reponse
func extractAlbumsFromLibrary(b []byte) []*Album {
	results := gjson.GetBytes(b, joinPaths(pSingleColumn, pTab0, pTabRendererContent, pSectionList, pContent0, pGridRendererItems))

	var albums []*Album
	results.ForEach(func(key, value gjson.Result) bool {
		if pl := extractAlbumFromLibrary(value); pl != nil {
			albums = append(albums, pl)
		}
		return true
	})

	return albums
}

// Parses a gjson.Result element into a Album. Return a pointer to it
// Expects one album element the brIDLikedAlbums endpoint JSON reponse
func extractAlbumFromLibrary(res gjson.Result) *Album {
	render := res.Get(pMusicTwoRow)
	if !render.Exists() {
		return nil
	}

	ab := &Album{
		Name: render.Get(joinPaths(pTitle, pRun0, pText)).String(),
		ID:   strings.TrimPrefix(render.Get(joinPaths(pTitle, pRun0, pNavEndpoint, pBrowseEnd, pBrowseID)).String(), "VL"),
	}

	ab.Author = extractUser(render.Get(joinPaths(pSubtitle, pRun2)))
	ab.Year, _ = strconv.Atoi(render.Get(joinPaths(pSubtitle, pRun4, pText)).String())

	return ab
}

// Parses res into a Album struct
// Expects the Album contained in browseId=VLPL... JSON response
func extractAlbum(res *gjson.Result) *Album {
	alb := &Album{}
	alb.Name = res.Get(pText).String()

	buf := res.Get(joinPaths(pNavEndpoint, pBrowseEnd, pBrowseID))
	if buf.Exists() {
		alb.ID = buf.String()
	}
	return alb
}
