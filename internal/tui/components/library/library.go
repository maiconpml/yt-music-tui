package library

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/maiconpml/heylisten/internal/tui/components/albums"
	"github.com/maiconpml/heylisten/internal/tui/components/playlists"
	"github.com/maiconpml/heylisten/internal/tui/components/tracks"
	"github.com/maiconpml/heylisten/internal/tui/keys"
	"github.com/maiconpml/heylisten/internal/tui/styles"
	"github.com/maiconpml/heylisten/pkg/goytmusic"
)

type libraryState int

const (
	stateRoot libraryState = iota
	stateDetails
)

type librarySection int

const (
	sectPlaylist = iota
	sectAlbum
)

type TracksLoadedMsg struct {
	Tracks []*goytmusic.Track
	Title  string
}

type ErrorMsg struct {
	Err error
}

type Model struct {
	client    *goytmusic.Client
	playlists playlists.Model
	albums    albums.Model
	tracks    tracks.Model
	state     libraryState
	section   librarySection
	width     int
	height    int
}

func New(client *goytmusic.Client) Model {
	return Model{
		client:    client,
		playlists: playlists.New(),
		albums:    albums.New(),
		tracks:    tracks.New(),
		state:     stateRoot,
		section:   sectPlaylist,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.playlists.Init(), m.albums.Init(), m.tracks.Init())
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Keys.Back):
			if m.state == stateDetails {
				m.state = stateRoot
				return m, nil
			}
		case key.Matches(msg, keys.LibraryKeys.NextSection):
			m.section = (m.section + 1) % 2
		case key.Matches(msg, keys.LibraryKeys.PrevSection):
			m.section = (m.section + 1) % 2
		}

	case playlists.PlaylistSelectedMsg:
		m.state = stateDetails
		id := msg.PlaylistID
		return m, func() tea.Msg {
			pl, err := m.client.Playlists.Get(&id)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			return TracksLoadedMsg{Tracks: pl.Tracks, Title: pl.Name}
		}

	case albums.AlbumSelectedMsg:
		m.state = stateDetails
		id := msg.AlbumID
		return m, func() tea.Msg {
			ab, err := m.client.Albums.Get(&id)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			return TracksLoadedMsg{Tracks: ab.Tracks, Title: ab.Name}
		}

	case TracksLoadedMsg:
		m.tracks.SetTracks(msg.Tracks, msg.Title)
		return m, nil
	}

	// Forward messages to active components
	var cmd tea.Cmd
	switch m.state {
	case stateRoot:
		switch msg.(type) {
		case playlists.PlaylistLoadedMsg:
			m.playlists, cmd = m.playlists.Update(msg)
			cmds = append(cmds, cmd)
		case albums.AlbumLoadedMsg:
			m.albums, cmd = m.albums.Update(msg)
			cmds = append(cmds, cmd)
		default:
			if m.section == sectPlaylist {
				m.playlists, cmd = m.playlists.Update(msg)
			} else {
				m.albums, cmd = m.albums.Update(msg)
			}
			cmds = append(cmds, cmd)
		}
	case stateDetails:
		m.tracks, cmd = m.tracks.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height - 1
	m.playlists.SetSize(width-2, height)
	m.albums.SetSize(width-2, height)
	m.tracks.SetSize(width-2, height)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var activeView string
	if m.state == stateRoot {
		if m.section == sectPlaylist {
			activeView = styles.RenderContainer(m.width, m.playlists.View(), "Playlists", styles.DimStyle.Render("Albums"))
		} else {
			activeView = styles.RenderContainer(m.width, m.albums.View(), styles.DimStyle.Render("Playlists"), "Albums")
		}
	} else {
		activeView = styles.RenderContainer(m.width, m.tracks.View(), m.tracks.Title())
	}

	return activeView
}

// Ptr returns a pointer to value v
func Ptr[T any](v T) *T {
	return &v
}
