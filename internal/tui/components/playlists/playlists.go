package playlists

import (
	"fmt"
	"io"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maiconpml/heylisten/internal/tui/styles"
	"github.com/maiconpml/heylisten/pkg/goytmusic"
)

type item struct {
	playlist *goytmusic.Playlist
}

func (i item) NTracks() int  { return i.playlist.NTracks }
func (i item) Title() string { return i.playlist.Name }
func (i item) Author() string {
	if i.playlist.Author != nil {
		return "by " + i.playlist.Author.Name
	}
	return "No author"
}
func (i item) FilterValue() string { return i.playlist.Name }

type playlistDelegate struct{}

func (d playlistDelegate) Height() int                               { return 1 }
func (d playlistDelegate) Spacing() int                              { return 0 }
func (d playlistDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d playlistDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	width := m.Width()
	if width <= 10 {
		return
	}
	avail := width - 2

	w1 := int(float64(avail) * 0.45)
	w2 := int(float64(avail) * 0.35)
	w3 := avail - w1 - w2

	var style, s1Style, s2Style, s3Style lipgloss.Style

	if index == m.Index() {
		style = styles.SelectedItemStyle
	} else {
		style = styles.BaseItemStyle
		s1Style, s2Style, s3Style = styles.NameStyle, styles.DimStyle, styles.DimStyle
	}

	renderCol := func(s lipgloss.Style, text string, colWidth int, paddingRight int) string {
		txt := styles.Truncate(text, colWidth-paddingRight)
		return s.
			Width(colWidth).
			Height(1).
			PaddingRight(paddingRight).
			Render(txt)
	}

	col1 := renderCol(s1Style, i.Title(), w1, 2)
	col2 := renderCol(s2Style, i.Author(), w2, 2)
	col3 := renderCol(s3Style, strconv.Itoa(i.NTracks())+" tracks", w3, 0)

	line := lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)

	fmt.Fprint(w, style.Width(width).Height(1).Render(line))
}

type Model struct {
	list   list.Model
	width  int
	height int
}

type PlaylistSelectedMsg struct{ PlaylistID string }

func New(playlists []*goytmusic.Playlist) Model {
	items := make([]list.Item, len(playlists))
	for i, p := range playlists {
		items[i] = item{playlist: p}
	}
	l := list.New(items, playlistDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return Model{list: l}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok && msg.String() == "enter" {
		if i, ok := m.list.SelectedItem().(item); ok {
			return m, func() tea.Msg {
				return PlaylistSelectedMsg{PlaylistID: i.playlist.ID}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) SetSize(w, h int) {
	m.width, m.height = w, h
	fw := styles.ContainerFrameWidth()
	m.list.SetSize(w-fw, h-2)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return m.list.View()
}
