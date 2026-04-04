package keys

import "github.com/charmbracelet/bubbles/key"

type LibraryKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Select      key.Binding
	NextSection key.Binding
	PrevSection key.Binding
	Play        key.Binding
	PlayRandom  key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view.
// Ordered by importance, ending with the toggle help command.
func (k LibraryKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.NextSection, k.PrevSection}
}

// FullHelp returns keybindings for the expanded help view.
func (k LibraryKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Select, k.NextSection, k.PrevSection},
	}
}

var LibraryKeys = LibraryKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	NextSection: key.NewBinding(
		key.WithKeys("]", ">"),
		key.WithHelp("]/>", "next section"),
	),
	PrevSection: key.NewBinding(
		key.WithKeys("[", "<"),
		key.WithHelp("[/<", "next section"),
	),
	Play: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "play playlist/music/album"),
	),
	PlayRandom: key.NewBinding(
		key.WithKeys("P"),
		key.WithHelp("P", "play playlist/album randomly"),
	),
}
