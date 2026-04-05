package goytmusic

import (
	"errors"

	"github.com/tidwall/gjson"
)

type TracksService service

type Track struct {
	Name               string
	Artists            []*User
	VideoID            *string
	PlaylistID         *string
	PlaylistSetVideoID *string
	Duration           string
	Album              *Album
}

// NextTracksByMusicInPlaylist returns the list of next tracks to be played
// when track videoID is selected in playlist playlistID having internal ID
// playlistSetVideoID. It also returns the continuation token to get next
// tracks. continuation can be passed to continue a previously retrieved queue
func (s *TracksService) NextTracksByMusicInPlaylist(videoID, playlistSetVideoID, playlistID, continuation *string) ([]*Track, string, error) {
	if s.client.isGuest {
		return nil, "", errors.New("Client is not authenticated")
	}
	body := struct {
		VideoID            string  `json:"videoId"`
		PlaylistSetVideoID string  `json:"playlistSetVideoId"`
		PlaylistID         string  `json:"playlistId"`
		Continuation       string  `json:"continuation"`
		Context            Context `json:"context"`
	}{*videoID, *playlistSetVideoID, *playlistID, *continuation, s.client.commonContext}

	return s.nextTracks(body)
}

// NextTracksByPlaylist returns the list of tracks to be played when playlist
// playlistID is selected. It also returns the continuation token to get next
// tracks Additional request parameters can be passed in params continuation
// can be passed to continue a previously retrieved queue.
func (s *TracksService) NextTracksByPlaylist(playlistID, continuation *string, random bool) ([]*Track, string, error) {
	if s.client.isGuest {
		return nil, "", errors.New("Client is not authenticated")
	}
	params := ""
	if random {
		params = "wAEB8gECKAE%3D"
	}
	body := struct {
		PlaylistID   string  `json:"playlistId"`
		Continuation string  `json:"continuation"`
		Params       string  `json:"params"`
		Context      Context `json:"context"`
	}{*playlistID, *continuation, params, s.client.commonContext}

	return s.nextTracks(body)
}

// nextTracks do a request to server next endpoint to retrieve next tracks
// to be played. It also returns the continuation token to the returned tracks.
// The request configuration is made by the caller by passing the request body.
func (s *TracksService) nextTracks(body any) ([]*Track, string, error) {
	u := "next"

	req, err := s.client.NewRequest("POST", u, body)
	if err != nil {
		return nil, "", err
	}

	respBody, _, err := s.client.Do(req)
	if err != nil {
		return nil, "", err
	}

	queue, continuation := extractTracksFromQueue(respBody)

	return queue, continuation, nil
}

// extractTracksFromQueue parses a JSON in []byte format into an array of Track
// pointers and a continuation token. Expects the queue response from the 'next' endpoint.
func extractTracksFromQueue(b []byte) ([]*Track, string) {
	results := gjson.GetBytes(b, joinPaths(pSingleColumnNextRts, pTabbedRenderer, pTab0, pTabRendererContent, pMusicQueueRenderer, pContent, pPlaylistPanelRenderer))

	resTracks := results.Get(pContents)
	var tracks []*Track
	resTracks.ForEach(func(key, value gjson.Result) bool {
		if tr := extractTrackFromQueue(value); tr != nil {
			tracks = append(tracks, tr)
		}
		return true
	})

	continuation := results.Get(pContinuation).String()
	return tracks, continuation
}

// extractTrackFromQueue parses res into a Track pointer
// Expects an element of the list returned by the 'next' endpoint
func extractTrackFromQueue(res gjson.Result) *Track {
	tr := &Track{}

	resTr := res.Get(joinPaths(pPlaylistPanelVideoWrapperRenderer, pPrimaryRenderer, pPlaylistPanelVideoRenderer))
	if !resTr.Exists() {
		resTr = res.Get(pPlaylistPanelVideoRenderer)
	}

	tr.Name = resTr.Get(joinPaths(pTitle, pRun0, pText)).String()
	buf := resTr.Get(joinPaths(pNavEndpoint, pWatchEnd, pVideoID))
	if buf.Exists() {
		tr.VideoID = Ptr(buf.String())
	}

	runs := resTr.Get(joinPaths(pLongByLineText, pRuns))

	runs.ForEach(func(key, value gjson.Result) bool {
		pageType := value.Get(joinPaths(pNavEndpoint, pBrowseEnd, pBrowseEndContextPageType)).String()

		switch pageType {
		case "MUSIC_PAGE_TYPE_ARTIST":
			if u := extractUser(value); u != nil {
				tr.Artists = append(tr.Artists, u)
			}
		case "MUSIC_PAGE_TYPE_ALBUM":
			tr.Album = extractAlbum(value)
		}

		return true
	})

	tr.Duration = resTr.Get(joinPaths(pLengthText, pRun0, pText)).String()

	return tr
}

// Parses res into a Track struct
// Expects the track contained in browseId=MPRE... endpoint JSON response
func extractTrackFromAlbum(res gjson.Result) *Track {
	tr := &Track{}

	tr.Name = res.Get(joinPaths(pRespListItem, pFlexColumn0, pRespListItemFlexColumn, pText, pRun0, pText)).String()
	buf := res.Get(joinPaths(pRespListItem, pOverlayRenderer, pContent, pMusicPlayButtonRenderer, pPlayNavEndpoint, pWatchEnd))
	if buf.Exists() {
		tr.VideoID = Ptr(buf.Get(pVideoID).String())
		tr.PlaylistID = Ptr(buf.Get(pPlaylistID).String())
		tr.PlaylistSetVideoID = Ptr(buf.Get(pPlaylistSetVideoID).String())
	}

	artists := res.Get(joinPaths(pRespListItem, pFlexColumn1, pRespListItemFlexColumn, pText, pRuns))
	artists.ForEach(func(key, value gjson.Result) bool {
		pageType := value.Get(joinPaths(pNavEndpoint, pBrowseEnd, pBrowseEndContextPageType)).String()
		if key.Int()%2 == 0 {
			if u := extractUser(value); u != nil && pageType == "MUSIC_PAGE_TYPE_ARTIST" {
				tr.Artists = append(tr.Artists, u)
			}
		}
		return true
	})

	return tr
}

// Parses res into a Track struct
// Expects the track contained in browseId=VLPL... endpoint JSON response
func extractTrack(res gjson.Result) *Track {
	tr := &Track{}

	tr.Name = res.Get(joinPaths(pRespListItem, pFlexColumn0, pRespListItemFlexColumn, pText, pRun0, pText)).String()
	buf := res.Get(joinPaths(pRespListItem, pOverlayRenderer, pContent, pMusicPlayButtonRenderer, pPlayNavEndpoint, pWatchEnd))
	if buf.Exists() {
		tr.VideoID = Ptr(buf.Get(pVideoID).String())
		tr.PlaylistID = Ptr(buf.Get(pPlaylistID).String())
		tr.PlaylistSetVideoID = Ptr(buf.Get(pPlaylistSetVideoID).String())
	}

	artists := res.Get(joinPaths(pRespListItem, pFlexColumn1, pRespListItemFlexColumn, pText, pRuns))
	artists.ForEach(func(key, value gjson.Result) bool {
		if key.Int()%2 == 0 {
			if u := extractUser(value); u != nil {
				tr.Artists = append(tr.Artists, u)
			}
		}
		return true
	})

	album := res.Get(joinPaths(pRespListItem, pFlexColumn2, pRespListItemFlexColumn, pText, pRun0))
	if album.Exists() {
		tr.Album = extractAlbum(album)
	}

	return tr
}
