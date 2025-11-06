package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ChartVersion represents a chart version option
type ChartVersion struct {
	Version     string
	AppVersion  string
	Description string
}

// UpgradeModel represents the upgrade view
type UpgradeModel struct {
	releaseName    string
	chartName      string
	currentVersion string
	versions       []ChartVersion
	cursor         int
	loading        bool
	upgrading      bool
	err            error
}

// NewUpgradeModel creates a new upgrade model
func NewUpgradeModel() *UpgradeModel {
	return &UpgradeModel{
		loading: false,
	}
}

// Init initializes the upgrade view
func (m *UpgradeModel) Init() tea.Cmd {
	return nil
}

// Update handles upgrade view messages
func (m *UpgradeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case upgradeVersionsMsg:
		m.versions = msg.versions
		m.loading = false
		return m, nil

	case upgradeActionMsg:
		m.upgrading = false
		if !msg.success {
			m.err = fmt.Errorf("%s", msg.message)
		}
		return m, nil

	case errMsg:
		m.err = error(msg)
		m.loading = false
		m.upgrading = false
		return m, nil
	}

	return m, nil
}

// View renders the upgrade view
func (m *UpgradeModel) View() string {
	if m.loading {
		return ProcessingStyle.Render("⟳ Loading available versions...")
	}

	if m.upgrading {
		return ProcessingStyle.Render("⟳ Upgrading release...")
	}

	if m.err != nil {
		return ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Upgrade: %s", m.releaseName)))
	b.WriteString("\n\n")

	// Current version info
	b.WriteString(CardStyle.Render(m.renderCurrentVersion()))
	b.WriteString("\n\n")

	// Available versions
	if len(m.versions) == 0 {
		b.WriteString(InfoMessageStyle.Render("No versions available"))
	} else {
		b.WriteString(m.renderVersionsList())
	}

	// Help text
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("↑/↓: navigate • enter: upgrade to selected version • esc: cancel"))

	return b.String()
}

// renderCurrentVersion renders current version info
func (m *UpgradeModel) renderCurrentVersion() string {
	var b strings.Builder

	b.WriteString("Current Installation:\n\n")
	b.WriteString(fmt.Sprintf("  Release: %s\n", InfoMessageStyle.Render(m.releaseName)))
	b.WriteString(fmt.Sprintf("  Chart:   %s\n", InfoMessageStyle.Render(m.chartName)))
	b.WriteString(fmt.Sprintf("  Version: %s\n", SuccessMessageStyle.Render(m.currentVersion)))

	return b.String()
}

// renderVersionsList renders the list of available versions
func (m *UpgradeModel) renderVersionsList() string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("Available Versions"))
	b.WriteString("\n\n")

	b.WriteString(HeaderStyle.Render(fmt.Sprintf("%-15s %-15s %s", "VERSION", "APP VERSION", "DESCRIPTION")))
	b.WriteString("\n")

	for i, ver := range m.versions {
		versionStr := Truncate(ver.Version, 15)
		appVersionStr := Truncate(ver.AppVersion, 15)
		descStr := Truncate(ver.Description, 50)

		// Highlight if this is the current version
		if ver.Version == m.currentVersion {
			versionStr = SuccessMessageStyle.Render(versionStr + " (current)")
		}

		row := fmt.Sprintf("%-15s %-15s %s", versionStr, appVersionStr, descStr)

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
func (m *UpgradeModel) MoveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveCursorDown moves the cursor down
func (m *UpgradeModel) MoveCursorDown() {
	if m.cursor < len(m.versions)-1 {
		m.cursor++
	}
}

// GetSelectedVersion returns the currently selected version
func (m *UpgradeModel) GetSelectedVersion() *ChartVersion {
	if len(m.versions) == 0 || m.cursor >= len(m.versions) {
		return nil
	}
	return &m.versions[m.cursor]
}

// SetRelease sets the release to upgrade
func (m *UpgradeModel) SetRelease(name, chart, version string) {
	m.releaseName = name
	m.chartName = chart
	m.currentVersion = version
	m.loading = true
	m.cursor = 0
	m.err = nil
}

// SetLoading sets the loading state
func (m *UpgradeModel) SetLoading(loading bool) {
	m.loading = loading
}

// SetUpgrading sets the upgrading state
func (m *UpgradeModel) SetUpgrading(upgrading bool) {
	m.upgrading = upgrading
}

// Reset resets the upgrade view
func (m *UpgradeModel) Reset() {
	m.releaseName = ""
	m.chartName = ""
	m.currentVersion = ""
	m.versions = nil
	m.cursor = 0
	m.loading = false
	m.upgrading = false
	m.err = nil
}
