package albums

import (
	"fmt"
	"io"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/maiconpml/heylisten/internal/tui/keys"
	"github.com/maiconpml/heylisten/internal/tui/styles"
	"github.com/maiconpml/heylisten/pkg/goytmusic"
)

type item struct {
	album *goytmusic.Album
}

func (i item) Year() int     { return i.album.Year }
func (i item) Title() string { return i.album.Name }
func (i item) Author() string {
	if i.album.Author != nil {
		return "by " + i.album.Author.Name
	}
	return "No author"
}
func (i item) FilterValue() string { return i.album.Name }

type albumDelegate struct{}

func (d albumDelegate) Height() int                               { return 1 }
func (d albumDelegate) Spacing() int                              { return 0 }
func (d albumDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d albumDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	width := m.Width()
	if width <= 10 {
		return
	}
	avail := width - 2

	w1 := int(float64(avail) * 0.6)
	w2 := int(float64(avail) * 0.3)
	w3 := avail - w1 - w2

	var style, s1Style, s2Style, s3Style lipgloss.Style

	if index == m.Index() {
		style = styles.SelectedItemStyle
		s3Style = styles.NameStyle.Align(lipgloss.Right)
	} else {
		style = styles.BaseItemStyle
		s1Style, s2Style, s3Style = styles.NameStyle, styles.DimStyle, styles.DimStyle.Align(lipgloss.Right)
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
	col3 := renderCol(s3Style, strconv.Itoa(i.Year()), w3, 0)

	line := lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)

	fmt.Fprint(w, style.Width(width).Height(1).Render(line))
}

type Model struct {
	list   list.Model
	width  int
	height int
}

type AlbumSelectedMsg struct{ AlbumID string }

type AlbumPlayedMsg struct {
	AlbumID string
	Random  bool
}

type AlbumLoadedMsg struct {
	Items []*goytmusic.Album
}

func New() Model {
	items := []list.Item{}
	l := list.New(items, albumDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)
	l.SetShowStatusBar(false)
	return Model{list: l}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.LibraryKeys.Select):
			if i, ok := m.list.SelectedItem().(item); ok {
				return m, func() tea.Msg {
					return AlbumSelectedMsg{AlbumID: i.album.ID}
				}
			}
		case key.Matches(msg, keys.LibraryKeys.Play):
			if i, ok := m.list.SelectedItem().(item); ok {
				return m, func() tea.Msg {
					return AlbumPlayedMsg{AlbumID: i.album.ID, Random: false}
				}
			}
		case key.Matches(msg, keys.LibraryKeys.PlayRandom):
			if i, ok := m.list.SelectedItem().(item); ok {
				return m, func() tea.Msg {
					return AlbumPlayedMsg{AlbumID: i.album.ID, Random: true}
				}
			}
		}
	case AlbumLoadedMsg:
		items := make([]list.Item, len(msg.Items))
		for i, a := range msg.Items {
			items[i] = item{album: a}
		}
		cmd := m.list.SetItems(items)
		return m, cmd
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
