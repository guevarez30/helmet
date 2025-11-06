package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/viewport"
)

// LogsModel represents the logs view for a pod
type LogsModel struct {
	podName  string
	logs     string
	loading  bool
	err      error
	viewport viewport.Model
	ready    bool
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
	}

	// Handle viewport scrolling
	if m.ready {
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

	// Help text
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: scroll • g/G: top/bottom • esc: back • q: quit"))

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
	if m.ready {
		m.viewport.SetContent("")
		m.viewport.GotoTop()
	}
}
