package tracks

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maiconpml/heylisten/internal/tui/styles"
	"github.com/maiconpml/heylisten/pkg/goytmusic"
)

type TrackSelectedMsg struct {
	Track *goytmusic.Track
	Index int
}

type item struct {
	track *goytmusic.Track
}

func (i item) Title() string { return i.track.Name }
func (i item) Artists() string {
	var artists []string
	for _, a := range i.track.Artists {
		artists = append(artists, a.Name)
	}
	return strings.Join(artists, ", ")
}

func (i item) AlbumInfo() string {
	info := ""
	if i.track.Album != nil {
		info = i.track.Album.Name
	}
	if i.track.Duration != "" {
		if info != "" {
			info += " • "
		}
		info += i.track.Duration
	}
	return info
}
func (i item) FilterValue() string { return i.track.Name }

type trackDelegate struct{}

func (d trackDelegate) Height() int                               { return 1 }
func (d trackDelegate) Spacing() int                              { return 0 }
func (d trackDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d trackDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
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
	w2 := int(float64(avail) * 0.30)
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
	col2 := renderCol(s2Style, i.Artists(), w2, 2)
	col3 := renderCol(s3Style, i.AlbumInfo(), w3, 0)

	line := lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)

	fmt.Fprint(w, style.Width(width).Height(1).Render(line))
}

type Model struct {
	list   list.Model
	width  int
	height int
	title  string
}

func New() Model {
	l := list.New(nil, trackDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)

	return Model{list: l}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetTracks(tracks []*goytmusic.Track, title string) {
	items := make([]list.Item, len(tracks))
	for i, tr := range tracks {
		items[i] = item{track: tr}
	}
	m.list.SetItems(items)
	m.list.SetSize(m.width-styles.ContainerFrameWidth(), m.height-2)
	m.title = title
}

func (m Model) Title() string {
	if m.title == "" {
		return "Músicas"
	}
	return m.title
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			index := m.list.Index()
			if i, ok := m.list.SelectedItem().(item); ok {
				return m, func() tea.Msg {
					return TrackSelectedMsg{Track: i.track, Index: index}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}
	return m.list.View()
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	fw := styles.ContainerFrameWidth()
	m.list.SetSize(width-fw, height-2)
}
