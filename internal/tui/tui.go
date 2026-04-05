package tui

import (
	"log/slog"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maiconpml/heylisten/internal/audio"
	"github.com/maiconpml/heylisten/internal/tui/components/albums"
	"github.com/maiconpml/heylisten/internal/tui/components/home"
	"github.com/maiconpml/heylisten/internal/tui/components/library"
	"github.com/maiconpml/heylisten/internal/tui/components/player"
	"github.com/maiconpml/heylisten/internal/tui/components/playlists"
	"github.com/maiconpml/heylisten/internal/tui/components/tracks"
	"github.com/maiconpml/heylisten/internal/tui/keys"
	"github.com/maiconpml/heylisten/internal/tui/styles"
	"github.com/maiconpml/heylisten/pkg/goytmusic"
	overlay "github.com/rmhubbert/bubbletea-overlay"
)

type tab int

const (
	tabHome tab = iota
	tabLibrary
	tabSearch // TODO:coming soon
)

type ErrorMsg struct {
	Err error
}

type Model struct {
	client     *goytmusic.Client
	tab        tab
	tabLibrary library.Model
	tabHome    home.Model
	player     player.Model
	help       help.Model
	viewport   viewport.Model
	keys       keys.KeyMap
	width      int
	height     int
}

func NewModel(client *goytmusic.Client) Model {
	return Model{
		client:     client,
		tabLibrary: library.New(client),
		tabHome:    home.New(),
		tab:        tabHome,
		player:     player.New(),
		help:       help.New(),
		viewport:   viewport.New(0, 0),
		keys:       keys.Keys,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.player.Init())
}

func (m *Model) updateSizes() {
	if m.width == 0 || m.height == 0 {
		return
	}

	paddingH := 4
	availWidth := m.width - paddingH

	tabsHeight := 1
	playerHeight := m.player.Height()

	footerHeight := lipgloss.Height(m.help.View(m.keys))

	availHeight := m.height - 3 - tabsHeight - playerHeight - footerHeight

	if availHeight < 0 {
		availHeight = 0
	}

	m.tabLibrary.SetSize(availWidth, availHeight)
	m.tabHome.SetSize(availWidth, availHeight)
}

func (m *Model) loadLibraryData() tea.Cmd {
	loadPl := func() tea.Msg {
		pls, err := m.client.Playlists.ListLiked()
		if err != nil {
			slog.Error("Failed to retrieved liked playlists", "err", err)
			return nil
		}
		return playlists.PlaylistLoadedMsg{Items: pls}
	}
	loadAb := func() tea.Msg {
		abs, err := m.client.Albums.ListLiked()
		if err != nil {
			slog.Error("Failed to retrieved liked albums", "err", err)
			return nil
		}
		return albums.AlbumLoadedMsg{Items: abs}
	}
	return tea.Batch(loadPl, loadAb)
}

func (m *Model) updateHelpViewport() {
	var sections []keys.HelpSection
	if m.tab == tabLibrary {
		merged := keys.MergedKeyMap{Global: m.keys, Library: keys.LibraryKeys}
		sections = merged.VerticalHelp()
	} else {
		sections = m.keys.VerticalHelp()
	}
	helpContent := keys.RenderVerticalHelp(sections)
	
	m.viewport.Width = min(m.width-10, 50)
	m.viewport.Height = min(lipgloss.Height(helpContent), m.height-10)
	m.viewport.SetContent(helpContent)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If help is shown, intercept keys to close it
		if m.help.ShowAll {
			switch {
			case key.Matches(msg, m.keys.Help),
				key.Matches(msg, m.keys.Back),
				key.Matches(msg, m.keys.Quit):
				m.help.ShowAll = false
				return m, nil
			}
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.NextTab):
			m.tab = (m.tab + 1) % 3
			if m.tab == 1 {
				cmds = append(cmds, m.loadLibraryData())
			}
		case key.Matches(msg, m.keys.PrevTab):
			m.tab--
			if m.tab < 0 {
				m.tab = 2
			} else if m.tab == 1 {
				cmds = append(cmds, m.loadLibraryData())
			}
		case key.Matches(msg, m.keys.TabHome):
			m.tab = 0
		case key.Matches(msg, m.keys.TabLibrary):
			m.tab = 1
			cmds = append(cmds, m.loadLibraryData())
		case key.Matches(msg, m.keys.TabSearch):
			m.tab = 2
		case key.Matches(msg, m.keys.PlayPause):
			return m, func() tea.Msg { return player.TogglePauseMsg{} }
		case key.Matches(msg, m.keys.NextTrack):
			return m, func() tea.Msg { return player.NextTrackMsg{} }
		case key.Matches(msg, m.keys.PrevTrack):
			return m, func() tea.Msg { return player.PrevTrackMsg{} }
		case key.Matches(msg, m.keys.VolUp):
			audio.IncrVolume()
		case key.Matches(msg, m.keys.VolDown):
			audio.DecrVolume()
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			if m.help.ShowAll {
				m.updateHelpViewport()
			}
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width - 4

		if m.help.ShowAll {
			m.updateHelpViewport()
		}

		availWidth := msg.Width - 4
		availHeight := msg.Height - 2

		var cmd tea.Cmd
		m.player, cmd = m.player.Update(tea.WindowSizeMsg{
			Width:  availWidth,
			Height: availHeight,
		})
		cmds = append(cmds, cmd)

		m.updateSizes()
		return m, tea.Batch(cmds...)

	case tracks.TrackSelectedMsg:
		return m, func() tea.Msg {
			cont := ""
			tracks, contin, err := m.client.Tracks.NextTracksByMusicInPlaylist(msg.Track.VideoID, msg.Track.PlaylistSetVideoID, msg.Track.PlaylistID, Ptr(cont))
			if err != nil {
				return ErrorMsg{Err: err}
			}
			return player.QueueLoadedMsg{Tracks: tracks, CurTrack: msg.Index, Continuation: contin}
		}

	case playlists.PlaylistPlayedMsg:
		return m, func() tea.Msg {
			cont := ""
			tracks, contin, err := m.client.Tracks.NextTracksByPlaylist(&msg.PlaylistID, Ptr(cont), msg.Random)
			if err != nil {
				return ErrorMsg{Err: err}
			}
			return player.QueueLoadedMsg{Tracks: tracks, CurTrack: 0, Continuation: contin}
		}
	}

	// Forward messages to active components
	var cmd tea.Cmd

	isPlayerTickMsg := false
	switch msg.(type) {
	case player.PlayerTickMsg, spinner.TickMsg, progress.FrameMsg:
		isPlayerTickMsg = true
	}

	if !isPlayerTickMsg {
		switch m.tab {
		case tabLibrary:
			m.tabLibrary, cmd = m.tabLibrary.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	m.player, cmd = m.player.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	availWidth := m.width - 4

	isHelpOpen := m.help.ShowAll

	tabsW := availWidth / 10
	if tabsW < 14 {
		tabsW = 14
	}

	tabStyle := lipgloss.NewStyle().Width(tabsW).Align(lipgloss.Center)
	selectedTabStyle := lipgloss.NewStyle().Width(tabsW).Align(lipgloss.Center).Background(lipgloss.Color("240"))
	homeTab := tabStyle.Render("[1]Home")
	libTab := tabStyle.Render("[2]Library")
	searchTab := tabStyle.Render("[3]Search")

	var activeView string
	switch m.tab {
	case tabHome:
		homeTab = selectedTabStyle.Render("[1]Home")
		activeView = m.tabHome.View()
	case tabLibrary:
		libTab = selectedTabStyle.Render("[2]Library")
		activeView = m.tabLibrary.View()
	case tabSearch:
		searchTab = selectedTabStyle.Render("[3]Search")
	}

	tabsView := lipgloss.JoinHorizontal(lipgloss.Left,
		homeTab,
		libTab,
		searchTab,
	)

	playerView := styles.RenderContainer(availWidth, m.player.View(), "Player")

	m.help.ShowAll = false
	var helpMap help.KeyMap = m.keys
	if m.tab == tabLibrary {
		helpMap = keys.MergedKeyMap{Global: m.keys, Library: keys.LibraryKeys}
	}
	m.help.Width = availWidth
	footerHelp := m.help.View(helpMap)

	m.help.ShowAll = isHelpOpen

	ui := lipgloss.JoinVertical(lipgloss.Left,
		tabsView,
		"",
		activeView,
		playerView,
		footerHelp,
	)

	background := lipgloss.NewStyle().
		Padding(0, 2).
		Width(m.width).
		Height(m.height).
		Render(ui)

	if isHelpOpen {
		helpHeader := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Italic(true).
			Render(" ↑/↓ Scroll • Esc/? Close")

		viewContent := lipgloss.JoinVertical(lipgloss.Center,
			helpHeader,
			"",
			m.viewport.View(),
		)

		modal := styles.ModalStyle.Render(viewContent)

		return overlay.Composite(
			modal,
			background,
			overlay.Center,
			overlay.Center,
			0, 0,
		)
	}
	return background
}

// Ptr returns a pointer to value v
func Ptr[T any](v T) *T {
	return &v
}
