package goytmusic

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL   = "https://music.youtube.com/youtubei/v1/"
	defaultUserAgent = "go-ytm/0.0.0"
	defaultOrigin    = "https://music.youtube.com"
)

// A Client abstracts the communication with Innertube API
type Client struct {
	// Base service for all other services to share
	common service

	commonContext Context

	// Services for each accessible resource of Innertube API
	Playlists *PlaylistsService
	Tracks    *TracksService

	isGuest bool

	// Http client to communicate with the API
	httpClient *http.Client
	// Base url for making requests.
	baseURL *url.URL
}

type service struct {
	client *Client
}

type Context struct {
	Client struct {
		ClientName    string `json:"clientName"`
		ClientVersion string `json:"clientVersion"`
		HL            string `json:"hl"`
		GL            string `json:"gl"`
	} `json:"client"`
}

// NewClient creats a client
func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	httpClient2 := *httpClient

	transport := httpClient2.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	// Ensures any request made by any client has User-Agent and X-Origin defined
	httpClient2.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req = req.Clone(req.Context())
		req.Header.Set("User-Agent", defaultUserAgent)
		req.Header.Set("X-Origin", defaultOrigin)

		query := req.URL.Query()
		query.Set("prettyPrint", "false")
		req.URL.RawQuery = query.Encode()

		return transport.RoundTrip(req)
	})

	c := &Client{httpClient: &httpClient2}

	c.baseURL, _ = url.Parse(defaultBaseURL)

	c.common.client = c

	c.commonContext.Client.ClientName = "WEB_REMIX"
	c.commonContext.Client.ClientVersion = "1.20240314.01.00"
	c.commonContext.Client.HL = "en"
	c.commonContext.Client.GL = "US"

	c.Playlists = (*PlaylistsService)(&c.common)
	c.Tracks = (*TracksService)(&c.common)
	c.Albums = (*AlbumService)(&c.common)

	return c
}

// WithAuthCookie returns a copy of client configured with authentication cookie
func (c *Client) WithAuthCookie(cookie string) *Client {
	sapisid := sapisidFromCookie(cookie)

	c2 := *c
	*c2.httpClient = *c.httpClient
	c2.isGuest = (cookie == "")
	transport := c2.httpClient.Transport

	c2.httpClient.Transport = roundTripperFunc(
		func(req *http.Request) (*http.Response, error) {
			req = req.Clone(req.Context())

			req.Header.Set("Cookie", cookie)

			if sapisid != "" {
				h := sapisidHash(sapisid)
				req.Header.Set("Authorization", fmt.Sprintf("SAPISIDHASH %s", h))
			}

			return transport.RoundTrip(req) // receiver client transport
		},
	)
	return &c2
}

// create SAPISIDHASH from the sapisid extracted from cookie.
// The Innertube API expects this SAPISIDHASH as auth identity
func sapisidHash(sapisid string) string {
	now := time.Now().Unix()

	concat := fmt.Sprintf("%d %s %s", now, sapisid, defaultOrigin)

	h := sha1.New()
	h.Write([]byte(concat))
	sha1Hash := fmt.Sprintf("%x", h.Sum(nil))

	return fmt.Sprintf("%d_%s", now, sha1Hash)
}

// Extract 3PAPISID from cookie
func sapisidFromCookie(cookie string) string {
	parts := strings.SplitSeq(cookie, ";")
	for part := range parts {
		part = strings.TrimSpace(part)
		if s, found := strings.CutPrefix(part, "__Secure-3PAPISID="); found {
			return s
		}
	}
	return ""
}

// NewRequest creates and returns a request given the request method, url and body.
// The url must be given as a string and relative to the baseURL.
func (c *Client) NewRequest(method, urlStr string, body any) (*http.Request, error) {
	u, err := c.baseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// Do sends an API request and returns the API response body as a byte slice.
// The method also closes the response body.
func (c *Client) Do(req *http.Request) ([]byte, *http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, resp, fmt.Errorf("api error: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp, err
	}

	return body, resp, nil
}

// BrowseBody returns a body for the browse endpoint requests.
// The body is composed by a browseID (see docs) and a default context.
func (c *Client) BrowseBody(browseID string) any {
	return struct {
		BrowseID string  `json:"browseId"`
		Context  Context `json:"context"`
	}{BrowseID: browseID, Context: c.commonContext}
}

// roundTripperFunc creates a RoundTripper (transport).
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return fn(r)
}

// Ptr returns a pointer to value v
func Ptr[T any](v T) *T {
	return &v
}
