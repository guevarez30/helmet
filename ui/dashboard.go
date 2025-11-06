package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"helm.sh/helm/v3/pkg/release"
)

// DashboardModel represents the dashboard view
type DashboardModel struct {
	releases      []*release.Release
	totalReleases int
	deployed      int
	failed        int
	pending       int
	loading       bool
	err           error
}

// NewDashboardModel creates a new dashboard model
func NewDashboardModel() *DashboardModel {
	return &DashboardModel{
		loading: true,
	}
}

// Init initializes the dashboard
func (m *DashboardModel) Init() tea.Cmd {
	return nil
}

// Update handles dashboard messages
func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case releasesMsg:
		m.releases = msg
		m.loading = false
		m.calculateStats()
		return m, nil
	case errMsg:
		m.err = error(msg)
		m.loading = false
		return m, nil
	}
	return m, nil
}

// View renders the dashboard
func (m *DashboardModel) View() string {
	if m.loading {
		return ProcessingStyle.Render("⟳ Loading dashboard...")
	}

	if m.err != nil {
		return ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Dashboard"))
	b.WriteString("\n\n")

	// Release statistics card
	releaseCard := m.renderReleaseCard()
	b.WriteString(releaseCard)

	// Quick start hint if no releases
	if m.totalReleases == 0 {
		b.WriteString("\n")
		b.WriteString(InfoMessageStyle.Render("No releases found. Press 's' to set up example charts or 'i' to install a chart."))
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Press tab to switch views • s: quick start • i: install chart • q: quit"))

	return b.String()
}

// renderReleaseCard renders the release statistics card
func (m *DashboardModel) renderReleaseCard() string {
	var b strings.Builder

	b.WriteString(CardTitleStyle.Render("Helm Releases"))
	b.WriteString("\n\n")

	// Total releases
	b.WriteString(fmt.Sprintf("Total Releases: %s\n",
		InfoMessageStyle.Render(fmt.Sprintf("%d", m.totalReleases))))

	// Deployed
	b.WriteString(fmt.Sprintf("  %s Deployed: %s\n",
		StatusIndicator("deployed"),
		DeployedStyle.Render(fmt.Sprintf("%d", m.deployed))))

	// Failed
	b.WriteString(fmt.Sprintf("  %s Failed: %s\n",
		StatusIndicator("failed"),
		FailedStyle.Render(fmt.Sprintf("%d", m.failed))))

	// Pending
	b.WriteString(fmt.Sprintf("  %s Pending: %s\n",
		StatusIndicator("pending"),
		PendingStyle.Render(fmt.Sprintf("%d", m.pending))))

	return CardStyle.Render(b.String())
}

// calculateStats calculates dashboard statistics
func (m *DashboardModel) calculateStats() {
	m.totalReleases = len(m.releases)
	m.deployed = 0
	m.failed = 0
	m.pending = 0

	for _, rel := range m.releases {
		switch rel.Info.Status {
		case release.StatusDeployed:
			m.deployed++
		case release.StatusFailed:
			m.failed++
		case release.StatusPendingInstall, release.StatusPendingUpgrade, release.StatusPendingRollback:
			m.pending++
		}
	}
}

// refresh returns a command to refresh dashboard data
func (m *DashboardModel) refresh() tea.Cmd {
	m.loading = true
	return nil
}

// SetReleases updates the releases data
func (m *DashboardModel) SetReleases(releases []*release.Release) {
	m.releases = releases
	m.calculateStats()
	m.loading = false
}
