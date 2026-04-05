package library

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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
	tracks    tracks.Model
	state     libraryState
	width     int
	height    int
}

func New(client *goytmusic.Client) Model {
	return Model{
		client:    client,
		playlists: playlists.New(),
		tracks:    tracks.New(),
		state:     stateRoot,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.playlists.Init(), m.tracks.Init())
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

	case TracksLoadedMsg:
		m.tracks.SetTracks(msg.Tracks, msg.Title)
		return m, nil
	}

	// Forward messages to active components
	var cmd tea.Cmd
	switch m.state {
	case stateRoot:
		m.playlists, cmd = m.playlists.Update(msg)
		cmds = append(cmds, cmd)
	case stateDetails:
		m.tracks, cmd = m.tracks.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.playlists.SetSize(width-2, height)
	m.tracks.SetSize(width-2, height)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	var activeView string
	if m.state == stateRoot {
		activeView = styles.RenderContainer("Minhas Playlists", m.width, m.playlists.View())
	} else {
		activeView = styles.RenderContainer(m.tracks.Title(), m.width, m.tracks.View())
	}

	return activeView
}

// Ptr returns a pointer to value v
func Ptr[T any](v T) *T {
	return &v
}
