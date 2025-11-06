package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"helm.sh/helm/v3/pkg/release"
)

// ReleasesModel represents the releases list view
type ReleasesModel struct {
	releases          []*release.Release
	cursor            int
	loading           bool
	err               error
	statusMsg         string
	actionInProgress  bool
}

// NewReleasesModel creates a new releases model
func NewReleasesModel() *ReleasesModel {
	return &ReleasesModel{
		loading: true,
	}
}

// Init initializes the releases view
func (m *ReleasesModel) Init() tea.Cmd {
	return nil
}

// Update handles releases view messages
func (m *ReleasesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case releasesMsg:
		m.releases = msg
		m.loading = false
		m.actionInProgress = false
		// Reset cursor if out of bounds
		if m.cursor >= len(m.releases) && len(m.releases) > 0 {
			m.cursor = len(m.releases) - 1
		}
		return m, nil

	case releaseActionMsg:
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

// View renders the releases list
func (m *ReleasesModel) View() string {
	if m.loading {
		return ProcessingStyle.Render("⟳ Loading releases...")
	}

	if m.err != nil {
		return ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Helm Releases (%d)", len(m.releases))))
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

	// Releases table
	if len(m.releases) == 0 {
		b.WriteString(InfoMessageStyle.Render("No releases found in this namespace"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(m.renderReleasesTable())
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: navigate • i: install • u: upgrade • d: delete • v: values • p: pods • tab: switch view • q: quit"))

	return b.String()
}

// renderReleasesTable renders the releases table
func (m *ReleasesModel) renderReleasesTable() string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("%-15s %-25s %-30s %-15s %-20s",
		"STATUS", "NAME", "CHART", "VERSION", "UPDATED")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	// Rows
	for i, rel := range m.releases {
		status := m.formatStatus(rel.Info.Status)
		name := Truncate(rel.Name, 25)
		chart := Truncate(rel.Chart.Metadata.Name, 30)
		version := Truncate(rel.Chart.Metadata.Version, 15)
		updated := m.formatTime(rel.Info.LastDeployed.Time)

		row := fmt.Sprintf("%-15s %-25s %-30s %-15s %-20s",
			status, name, chart, version, updated)

		if i == m.cursor {
			b.WriteString(SelectedItemStyle.Render(row))
		} else {
			b.WriteString(UnselectedItemStyle.Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// formatStatus formats the release status with color and indicator
func (m *ReleasesModel) formatStatus(status release.Status) string {
	statusStr := string(status)
	indicator := StatusIndicator(statusStr)
	return fmt.Sprintf("%s %s", indicator, Truncate(statusStr, 10))
}

// formatTime formats a timestamp to relative time
func (m *ReleasesModel) formatTime(t time.Time) string {
	duration := time.Since(t.UTC())

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%ds ago", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%dm ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(duration.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(duration.Hours()/24))
	}
}

// MoveCursorUp moves the cursor up
func (m *ReleasesModel) MoveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveCursorDown moves the cursor down
func (m *ReleasesModel) MoveCursorDown() {
	if m.cursor < len(m.releases)-1 {
		m.cursor++
	}
}

// GetSelectedRelease returns the currently selected release
func (m *ReleasesModel) GetSelectedRelease() *release.Release {
	if len(m.releases) == 0 || m.cursor >= len(m.releases) {
		return nil
	}
	return m.releases[m.cursor]
}

// SetReleases updates the releases data
func (m *ReleasesModel) SetReleases(releases []*release.Release) {
	m.releases = releases
	m.loading = false
}

// SetActionInProgress sets the action in progress flag
func (m *ReleasesModel) SetActionInProgress(inProgress bool) {
	m.actionInProgress = inProgress
}
