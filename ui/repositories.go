package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"helm.sh/helm/v3/pkg/repo"
)

// RepositoriesModel represents the repositories list view
type RepositoriesModel struct {
	repositories     []*repo.Entry
	cursor           int
	loading          bool
	err              error
	statusMsg        string
	actionInProgress bool
}

// NewRepositoriesModel creates a new repositories model
func NewRepositoriesModel() *RepositoriesModel {
	return &RepositoriesModel{
		loading: true,
	}
}

// Init initializes the repositories view
func (m *RepositoriesModel) Init() tea.Cmd {
	return nil
}

// Update handles repositories view messages
func (m *RepositoriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repositoriesMsg:
		m.repositories = msg
		m.loading = false
		m.actionInProgress = false
		// Reset cursor if out of bounds
		if m.cursor >= len(m.repositories) && len(m.repositories) > 0 {
			m.cursor = len(m.repositories) - 1
		}
		return m, nil

	case repoActionMsg:
		m.actionInProgress = false
		if msg.success {
			m.statusMsg = msg.message
			// Auto-clear status after 2 seconds
			return m, tea.Sequence(
				tea.Tick(2*time.Second, func(time.Time) tea.Msg {
					return clearStatusMsg{}
				}),
			)
		} else {
			m.err = fmt.Errorf("%s", msg.message)
		}
		return m, nil

	case clearStatusMsg:
		m.statusMsg = ""
		return m, nil

	case errMsg:
		m.err = error(msg)
		m.loading = false
		m.actionInProgress = false
		return m, nil
	}

	return m, nil
}

// View renders the repositories list
func (m *RepositoriesModel) View() string {
	if m.loading {
		return ProcessingStyle.Render("⟳ Loading repositories...")
	}

	if m.err != nil {
		return ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Helm Repositories (%d)", len(m.repositories))))
	b.WriteString("\n\n")

	// Status message
	if m.statusMsg != "" {
		b.WriteString(SuccessMessageStyle.Render("✓ " + m.statusMsg))
		b.WriteString("\n\n")
	}

	// Action in progress indicator
	if m.actionInProgress {
		b.WriteString(ProcessingStyle.Render("⟳ Processing..."))
		b.WriteString("\n\n")
	}

	// Repositories table
	if len(m.repositories) == 0 {
		b.WriteString(InfoMessageStyle.Render("No repositories configured"))
		b.WriteString("\n\n")
		b.WriteString(HelpStyle.Render("Press 'a' to add a repository"))
	} else {
		b.WriteString(m.renderRepositoriesTable())
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: navigate • a: add repo • r: remove • U: update • tab: switch view • q: quit"))

	return b.String()
}

// renderRepositoriesTable renders the repositories table
func (m *RepositoriesModel) renderRepositoriesTable() string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("%-30s %-60s",
		"NAME", "URL")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	// Rows
	for i, r := range m.repositories {
		name := Truncate(r.Name, 30)
		url := Truncate(r.URL, 60)

		row := fmt.Sprintf("%-30s %-60s", name, url)

		if i == m.cursor {
			b.WriteString(SelectedItemStyle.Render(row))
		} else {
			b.WriteString(UnselectedItemStyle.Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// MoveCursorUp moves the cursor up
func (m *RepositoriesModel) MoveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveCursorDown moves the cursor down
func (m *RepositoriesModel) MoveCursorDown() {
	if m.cursor < len(m.repositories)-1 {
		m.cursor++
	}
}

// GetSelectedRepository returns the currently selected repository
func (m *RepositoriesModel) GetSelectedRepository() *repo.Entry {
	if len(m.repositories) == 0 || m.cursor >= len(m.repositories) {
		return nil
	}
	return m.repositories[m.cursor]
}

// SetRepositories updates the repositories data
func (m *RepositoriesModel) SetRepositories(repositories []*repo.Entry) {
	m.repositories = repositories
	m.loading = false
}

// SetActionInProgress sets the action in progress flag
func (m *RepositoriesModel) SetActionInProgress(inProgress bool) {
	m.actionInProgress = inProgress
}
