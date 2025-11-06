package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
)

// LogsModel represents the logs view for a pod
type LogsModel struct {
	podName       string
	logs          string
	loading       bool
	err           error
	viewport      viewport.Model
	ready         bool
	searchMode    bool
	searchQuery   string
	searchMatches []int // line numbers of matches
	currentMatch  int   // index in searchMatches
}

// NewLogsModel creates a new logs model
func NewLogsModel() *LogsModel {
	return &LogsModel{
		loading: false,
	}
}

// Init initializes the logs view
func (m *LogsModel) Init() tea.Cmd {
	return nil
}

// Update handles logs view messages
func (m *LogsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case logsMsg:
		m.logs = msg.logs
		m.podName = msg.podName
		m.loading = false
		if m.ready {
			m.viewport.SetContent(msg.logs)
		}
		return m, nil

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-8)
			m.viewport.YPosition = 5
			m.ready = true
			if m.logs != "" {
				m.viewport.SetContent(m.logs)
			}
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 8
		}
		return m, nil

	case errMsg:
		m.err = error(msg)
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		// Handle search mode input
		if m.searchMode {
			switch msg.String() {
			case "enter":
				if m.searchQuery != "" {
					m.executeSearch()
				}
				m.searchMode = false
				return m, nil
			case "esc":
				m.searchMode = false
				m.searchQuery = ""
				return m, nil
			case "backspace":
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
				}
				return m, nil
			default:
				// Add character to search query
				if len(msg.String()) == 1 {
					m.searchQuery += msg.String()
				}
				return m, nil
			}
		}

		// Normal mode key handling
		switch msg.String() {
		case "/":
			m.searchMode = true
			m.searchQuery = ""
			m.clearSearch()
			return m, nil
		case "n":
			m.nextMatch()
			return m, nil
		case "N":
			m.prevMatch()
			return m, nil
		}
	}

	// Handle viewport scrolling (only when not in search mode)
	if m.ready && !m.searchMode {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

// View renders the logs view
func (m *LogsModel) View() string {
	if m.loading {
		return ProcessingStyle.Render("⟳ Loading logs...")
	}

	if m.err != nil {
		return ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Logs: %s", m.podName)))
	b.WriteString("\n\n")

	if !m.ready {
		b.WriteString(InfoMessageStyle.Render("Initializing viewport..."))
		b.WriteString("\n")
	} else {
		// Viewport with logs
		b.WriteString(m.viewport.View())
		b.WriteString("\n")
	}

	// Search bar or search status
	b.WriteString("\n")
	if m.searchMode {
		b.WriteString(InfoMessageStyle.Render(fmt.Sprintf("Search: %s_", m.searchQuery)))
	} else if m.searchQuery != "" && len(m.searchMatches) == 0 {
		b.WriteString(ErrorMessageStyle.Render("No matches found"))
	} else if len(m.searchMatches) > 0 {
		b.WriteString(SuccessMessageStyle.Render(fmt.Sprintf("Match %d/%d", m.currentMatch+1, len(m.searchMatches))))
	}

	// Help text
	b.WriteString("\n")
	if m.searchMode {
		b.WriteString(HelpStyle.Render("enter: search • esc: cancel"))
	} else {
		b.WriteString(HelpStyle.Render("↑/↓: scroll • /: search • n/N: next/prev • g/G: top/bottom • esc: back • q: quit"))
	}

	return b.String()
}

// SetLoading sets the loading state
func (m *LogsModel) SetLoading(loading bool) {
	m.loading = loading
}

// SetLogs sets the logs data
func (m *LogsModel) SetLogs(podName, logs string) {
	m.podName = podName
	m.logs = logs
	m.loading = false
	if m.ready {
		m.viewport.SetContent(logs)
	}
}

// Reset resets the logs model
func (m *LogsModel) Reset() {
	m.podName = ""
	m.logs = ""
	m.loading = false
	m.err = nil
	m.searchMode = false
	m.searchQuery = ""
	m.searchMatches = nil
	m.currentMatch = 0
	if m.ready {
		m.viewport.SetContent("")
		m.viewport.GotoTop()
	}
}

// executeSearch searches for the query in the logs
func (m *LogsModel) executeSearch() {
	if m.searchQuery == "" {
		return
	}

	m.searchMatches = nil
	m.currentMatch = 0

	lines := strings.Split(m.logs, "\n")
	query := strings.ToLower(m.searchQuery)

	for i, line := range lines {
		if strings.Contains(strings.ToLower(line), query) {
			m.searchMatches = append(m.searchMatches, i)
		}
	}

	// Update viewport with highlighted content
	m.updateHighlightedContent()

	// Jump to first match
	if len(m.searchMatches) > 0 {
		m.jumpToLine(m.searchMatches[0])
	}
}

// nextMatch jumps to the next search match
func (m *LogsModel) nextMatch() {
	if len(m.searchMatches) == 0 {
		return
	}

	m.currentMatch = (m.currentMatch + 1) % len(m.searchMatches)
	m.updateHighlightedContent()
	m.jumpToLine(m.searchMatches[m.currentMatch])
}

// prevMatch jumps to the previous search match
func (m *LogsModel) prevMatch() {
	if len(m.searchMatches) == 0 {
		return
	}

	m.currentMatch--
	if m.currentMatch < 0 {
		m.currentMatch = len(m.searchMatches) - 1
	}
	m.updateHighlightedContent()
	m.jumpToLine(m.searchMatches[m.currentMatch])
}

// jumpToLine scrolls the viewport to show the specified line
func (m *LogsModel) jumpToLine(lineNum int) {
	if !m.ready {
		return
	}

	// Calculate the offset to center the line in the viewport
	targetOffset := lineNum - (m.viewport.Height / 2)
	if targetOffset < 0 {
		targetOffset = 0
	}

	// Set the viewport's YOffset to scroll to the line
	m.viewport.SetYOffset(targetOffset)
}

// updateHighlightedContent updates the viewport content with search highlights
func (m *LogsModel) updateHighlightedContent() {
	if !m.ready || m.searchQuery == "" || len(m.searchMatches) == 0 {
		return
	}

	lines := strings.Split(m.logs, "\n")
	query := strings.ToLower(m.searchQuery)
	currentMatchLine := -1
	if m.currentMatch < len(m.searchMatches) {
		currentMatchLine = m.searchMatches[m.currentMatch]
	}

	var highlightedLines []string
	for i, line := range lines {
		lowerLine := strings.ToLower(line)
		if !strings.Contains(lowerLine, query) {
			highlightedLines = append(highlightedLines, line)
			continue
		}

		// Highlight all occurrences in this line
		highlightedLine := ""
		remaining := line
		remainingLower := lowerLine

		for {
			idx := strings.Index(remainingLower, query)
			if idx == -1 {
				highlightedLine += remaining
				break
			}

			// Add text before match
			highlightedLine += remaining[:idx]

			// Highlight the match - use different style for current match
			matchText := remaining[idx : idx+len(m.searchQuery)]
			if i == currentMatchLine {
				// Current match - brighter highlight
				highlightedLine += SuccessMessageStyle.Render(matchText)
			} else {
				// Other matches - standard highlight
				highlightedLine += HighlightStyle.Render(matchText)
			}

			// Continue with rest of line
			remaining = remaining[idx+len(m.searchQuery):]
			remainingLower = remainingLower[idx+len(m.searchQuery):]
		}

		highlightedLines = append(highlightedLines, highlightedLine)
	}

	m.viewport.SetContent(strings.Join(highlightedLines, "\n"))
}

// clearSearch clears the search state and restores original content
func (m *LogsModel) clearSearch() {
	m.searchMatches = nil
	m.currentMatch = 0
	if m.ready && m.logs != "" {
		m.viewport.SetContent(m.logs)
	}
}
