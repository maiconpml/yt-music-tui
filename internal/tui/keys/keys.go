package keys

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type KeyMap struct {
	PlayPause  key.Binding
	NextTrack  key.Binding
	PrevTrack  key.Binding
	NextTab    key.Binding
	TabHome    key.Binding
	TabLibrary key.Binding
	TabSearch  key.Binding
	PrevTab    key.Binding
	VolUp      key.Binding
	VolDown    key.Binding
	Help       key.Binding
	Quit       key.Binding
	Back       key.Binding
}

type HelpSection struct {
	Title    string
	Bindings []key.Binding
}

func RenderVerticalHelp(sections []HelpSection) string {
	var sb strings.Builder

	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("62")).Bold(true).MarginBottom(1)
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true).Width(12).Align(lipgloss.Right)
	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).PaddingLeft(1)

	for i, sec := range sections {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(titleStyle.Render(sec.Title) + "\n")
		for _, b := range sec.Bindings {
			if !b.Enabled() {
				continue
			}
			h := b.Help()
			line := lipgloss.JoinHorizontal(lipgloss.Top, keyStyle.Render(h.Key), descStyle.Render(h.Desc))
			sb.WriteString(line + "\n")
		}
	}

	return sb.String()
}

// ShortHelp returns keybindings to be shown in the mini help view.
// Ordered by importance, ending with the toggle help command.
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.NextTab, k.PlayPause, k.NextTrack, k.Back, k.Quit}
}

// FullHelp returns keybindings for the expanded help view.
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.NextTab, k.PrevTab, k.Back, k.TabHome, k.TabLibrary, k.TabSearch},
		{k.PlayPause, k.NextTrack, k.PrevTrack, k.VolUp, k.VolDown, k.Help, k.Quit},
	}
}

func (k KeyMap) VerticalHelp() []HelpSection {
	return []HelpSection{
		{
			Title:    "Navigation",
			Bindings: []key.Binding{k.NextTab, k.PrevTab, k.TabHome, k.TabLibrary, k.TabSearch, k.Back},
		},
		{
			Title:    "Playback",
			Bindings: []key.Binding{k.PlayPause, k.NextTrack, k.PrevTrack, k.VolUp, k.VolDown},
		},
		{
			Title:    "App",
			Bindings: []key.Binding{k.Help, k.Quit},
		},
	}
}

type MergedKeyMap struct {
	Global  KeyMap
	Library LibraryKeyMap
}

func (k MergedKeyMap) ShortHelp() []key.Binding {
	// Combine most important global keys and library keys
	return []key.Binding{k.Global.Help, k.Global.NextTab, k.Global.PlayPause, k.Library.Select, k.Library.Play, k.Global.Back, k.Global.Quit}
}

func (k MergedKeyMap) FullHelp() [][]key.Binding {
	full := k.Global.FullHelp()
	full = append(full, k.Library.FullHelp()...)
	return full
}

func (k MergedKeyMap) VerticalHelp() []HelpSection {
	sections := k.Global.VerticalHelp()

	libSection := HelpSection{
		Title:    "Library",
		Bindings: []key.Binding{k.Library.Up, k.Library.Down, k.Library.Select, k.Library.NextSection, k.Library.PrevSection, k.Library.Play, k.Library.PlayRandom},
	}

	newSections := append([]HelpSection{}, sections[:2]...)
	newSections = append(newSections, libSection)
	newSections = append(newSections, sections[2:]...)

	return newSections
}

var Keys = KeyMap{
	TabHome: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "go to home")),
	TabLibrary: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "go to library")),
	TabSearch: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "go to search")),
	NextTab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab")),
	PrevTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab")),
	PlayPause: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "play/pause"),
	),
	NextTrack: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "next track"),
	),
	PrevTrack: key.NewBinding(
		key.WithKeys("N"),
		key.WithHelp("N", "prev track"),
	),
	VolUp: key.NewBinding(
		key.WithKeys("+"),
		key.WithHelp("+", "vol up"),
	),
	VolDown: key.NewBinding(
		key.WithKeys("-"),
		key.WithHelp("-", "vol down"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "keybinds"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc", "backspace"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
