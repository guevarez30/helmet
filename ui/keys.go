package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines keybindings for the application
type KeyMap struct {
	Up           key.Binding
	Down         key.Binding
	Left         key.Binding
	Right        key.Binding
	Tab          key.Binding
	Enter        key.Binding
	Escape       key.Binding
	Quit         key.Binding
	Refresh      key.Binding
	Start        key.Binding
	Stop         key.Binding
	Restart      key.Binding
	Delete       key.Binding
	Upgrade      key.Binding
	Install      key.Binding
	Logs         key.Binding
	History      key.Binding
	Rollback     key.Binding
	GetValues    key.Binding
	Inspect      key.Binding
	Search       key.Binding
	Add          key.Binding
	Update       key.Binding
	SwitchCtx    key.Binding
	SwitchNs     key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "move right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "switch view"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "back/cancel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Start: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "start"),
		),
		Stop: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "stop"),
		),
		Restart: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "restart/rollback"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Upgrade: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "upgrade"),
		),
		Install: key.NewBinding(
			key.WithKeys("i"),
			key.WithHelp("i", "install/inspect"),
		),
		Logs: key.NewBinding(
			key.WithKeys("L"),
			key.WithHelp("L", "logs"),
		),
		History: key.NewBinding(
			key.WithKeys("H"),
			key.WithHelp("H", "history"),
		),
		Rollback: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "rollback"),
		),
		GetValues: key.NewBinding(
			key.WithKeys("v"),
			key.WithHelp("v", "view values"),
		),
		Inspect: key.NewBinding(
			key.WithKeys("I"),
			key.WithHelp("I", "inspect"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Add: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "add"),
		),
		Update: key.NewBinding(
			key.WithKeys("U"),
			key.WithHelp("U", "update"),
		),
		SwitchCtx: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "switch context"),
		),
		SwitchNs: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "switch namespace"),
		),
	}
}
