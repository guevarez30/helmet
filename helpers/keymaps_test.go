package helpers

import (
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/stretchr/testify/assert"
)

// TestCommonKeys verifies that the CommonKeys are correctly initialized.
func TestCommonKeys(t *testing.T) {
	expectedTabNav := key.NewBinding(key.WithKeys("tab", "shift+tab"), key.WithHelp("tab/S-tab", "Next/prev tab"))
	expectedHistoryNav := key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "History back/fwd"))
	expectedCommand := key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "Command"))
	expectedHelp := key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "Help"))
	expectedQuit := key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Quit"))

	assert.Equal(t, expectedTabNav, CommonKeys.TabNav, "TabNav keybinding should match")
	assert.Equal(t, expectedHistoryNav, CommonKeys.HistoryNav, "HistoryNav keybinding should match")
	assert.Equal(t, expectedCommand, CommonKeys.Command, "Command keybinding should match")
	assert.Equal(t, expectedHelp, CommonKeys.Help, "Help keybinding should match")
	assert.Equal(t, expectedQuit, CommonKeys.Quit, "Quit keybinding should match")
}

// TestShortHelp verifies that ShortHelp returns the correct keybindings for keyMap.
func TestKeyMapShortHelp(t *testing.T) {
	expected := []key.Binding{
		CommonKeys.TabNav,
		CommonKeys.Command,
		CommonKeys.HistoryNav,
		CommonKeys.Help,
		CommonKeys.Quit,
	}

	shortHelp := CommonKeys.ShortHelp()
	assert.Equal(t, expected, shortHelp, "ShortHelp should return the correct keybindings")
}

// TestSuggestionInputKeyMap verifies that SuggestionInputKeyMap is correctly initialized.
func TestSuggestionInputKeyMap(t *testing.T) {
	expectedAcceptSuggestion := key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Accept suggestion"))
	expectedNextSuggestion := key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down/ctrl+n", "Next suggestion"))
	expectedPrevSuggestion := key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up/ctrl+p", "Previous suggestion"))

	assert.Equal(t, expectedAcceptSuggestion, SuggestionInputKeyMap.AcceptSuggestion, "AcceptSuggestion keybinding should match")
	assert.Equal(t, expectedNextSuggestion, SuggestionInputKeyMap.NextSuggestion, "NextSuggestion keybinding should match")
	assert.Equal(t, expectedPrevSuggestion, SuggestionInputKeyMap.PrevSuggestion, "PrevSuggestion keybinding should match")
}

// TestSuggestionKeyMapShortHelp verifies that ShortHelp returns the correct keybindings for SuggestionKeyMap.
func TestSuggestionKeyMapShortHelp(t *testing.T) {
	expected := []key.Binding{
		SuggestionInputKeyMap.AcceptSuggestion,
		SuggestionInputKeyMap.NextSuggestion,
		SuggestionInputKeyMap.PrevSuggestion,
	}

	shortHelp := SuggestionInputKeyMap.ShortHelp()
	assert.Equal(t, expected, shortHelp, "ShortHelp should return the correct keybindings")
}
