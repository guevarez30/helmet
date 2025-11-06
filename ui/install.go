package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/guevarez30/helmet/helm"
)

// InstallModel represents the chart installation form
type InstallModel struct {
	releaseName  textinput.Model
	chartPath    textinput.Model
	namespace    textinput.Model
	focusIndex   int
	inputs       []textinput.Model
	err          error
	installing   bool
	localCharts  []helm.LocalChart
	loadingCharts bool
}

// NewInstallModel creates a new install model
func NewInstallModel() *InstallModel {
	m := &InstallModel{
		releaseName: textinput.New(),
		chartPath:   textinput.New(),
		namespace:   textinput.New(),
		focusIndex:  0,
	}

	m.releaseName.Placeholder = "my-release"
	m.releaseName.Focus()
	m.releaseName.CharLimit = 50
	m.releaseName.Width = 50
	m.releaseName.Prompt = "Release Name: "

	m.chartPath.Placeholder = "./mychart or bitnami/nginx"
	m.chartPath.CharLimit = 200
	m.chartPath.Width = 50
	m.chartPath.Prompt = "Chart Path:   "

	m.namespace.Placeholder = "default"
	m.namespace.CharLimit = 50
	m.namespace.Width = 50
	m.namespace.Prompt = "Namespace:    "

	m.inputs = []textinput.Model{m.releaseName, m.chartPath, m.namespace}

	return m
}

// Init initializes the install form
func (m *InstallModel) Init() tea.Cmd {
	// Discover local charts
	return tea.Batch(
		textinput.Blink,
		m.discoverCharts(),
	)
}

// discoverCharts searches for local Helm charts
func (m *InstallModel) discoverCharts() tea.Cmd {
	m.loadingCharts = true
	return func() tea.Msg {
		charts, err := helm.DiscoverLocalCharts(".")
		if err != nil {
			return localChartsMsg{charts: nil}
		}
		return localChartsMsg{charts: charts}
	}
}

type localChartsMsg struct {
	charts []helm.LocalChart
}

// Update handles install form messages
func (m *InstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case localChartsMsg:
		m.localCharts = msg.charts
		m.loadingCharts = false
		return m, nil

	case installActionMsg:
		m.installing = false
		if !msg.success {
			m.err = fmt.Errorf("%s", msg.message)
		}
		return m, nil

	case errMsg:
		m.err = error(msg)
		m.installing = false
		return m, nil
	}

	// Handle input updates
	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)

	// Update the actual fields
	m.releaseName = m.inputs[0]
	m.chartPath = m.inputs[1]
	m.namespace = m.inputs[2]

	return m, cmd
}

// View renders the install form
func (m *InstallModel) View() string {
	if m.installing {
		return ProcessingStyle.Render("⟳ Installing chart...")
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Install Helm Chart"))
	b.WriteString("\n\n")

	// Error message
	if m.err != nil {
		b.WriteString(ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	// Form fields
	b.WriteString(CardStyle.Render(m.renderForm()))

	// Help text
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("tab/shift+tab: switch field • enter: install • esc: cancel"))

	return b.String()
}

// renderForm renders the form fields
func (m *InstallModel) renderForm() string {
	var b strings.Builder

	b.WriteString("Fill in the details to install a Helm chart:\n\n")

	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")

	// Show discovered local charts
	if m.loadingCharts {
		b.WriteString("Discovering local charts...\n")
	} else if len(m.localCharts) > 0 {
		b.WriteString("Discovered Local Charts:\n")
		for _, chart := range m.localCharts {
			b.WriteString(fmt.Sprintf("  • %s (%s)\n", chart.Name, chart.Path))
		}
		b.WriteString("\n")
	}

	b.WriteString("Examples:\n")
	b.WriteString("  Local chart:  ./mychart or ../charts/mychart\n")
	b.WriteString("  Repo chart:   bitnami/nginx or stable/redis\n")

	return b.String()
}

// MoveFocusNext moves focus to the next input
func (m *InstallModel) MoveFocusNext() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
	m.inputs[m.focusIndex].Focus()
}

// MoveFocusPrev moves focus to the previous input
func (m *InstallModel) MoveFocusPrev() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs) - 1
	}
	m.inputs[m.focusIndex].Focus()
}

// GetReleaseName returns the release name
func (m *InstallModel) GetReleaseName() string {
	name := m.releaseName.Value()
	if name == "" {
		return m.releaseName.Placeholder
	}
	return name
}

// GetChartPath returns the chart path
func (m *InstallModel) GetChartPath() string {
	return m.chartPath.Value()
}

// GetNamespace returns the namespace
func (m *InstallModel) GetNamespace() string {
	ns := m.namespace.Value()
	if ns == "" {
		return m.namespace.Placeholder
	}
	return ns
}

// SetInstalling sets the installing state
func (m *InstallModel) SetInstalling(installing bool) {
	m.installing = installing
}

// Reset resets the form
func (m *InstallModel) Reset() {
	m.releaseName.SetValue("")
	m.chartPath.SetValue("")
	m.namespace.SetValue("")
	m.err = nil
	m.focusIndex = 0
	m.inputs[0].Focus()
}
