package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// ChartCatalogItem represents a chart in the catalog
type ChartCatalogItem struct {
	Name        string
	Chart       string
	Description string
	RepoName    string
	RepoURL     string
}

// ChartCatalogModel represents the chart catalog browser
type ChartCatalogModel struct {
	charts     []ChartCatalogItem
	cursor     int
	selected   map[int]bool
	installing bool
	err        error
}

// NewChartCatalogModel creates a new chart catalog model
func NewChartCatalogModel() *ChartCatalogModel {
	return &ChartCatalogModel{
		charts:   getPopularCharts(),
		selected: make(map[int]bool),
	}
}

// getPopularCharts returns a list of popular Helm charts
func getPopularCharts() []ChartCatalogItem {
	return []ChartCatalogItem{
		// Bitnami charts
		{
			Name:        "NGINX",
			Chart:       "bitnami/nginx",
			Description: "Web server and reverse proxy",
			RepoName:    "bitnami",
			RepoURL:     "https://charts.bitnami.com/bitnami",
		},
		{
			Name:        "Redis",
			Chart:       "bitnami/redis",
			Description: "In-memory data store and cache",
			RepoName:    "bitnami",
			RepoURL:     "https://charts.bitnami.com/bitnami",
		},
		{
			Name:        "PostgreSQL",
			Chart:       "bitnami/postgresql",
			Description: "Relational database",
			RepoName:    "bitnami",
			RepoURL:     "https://charts.bitnami.com/bitnami",
		},
		{
			Name:        "MongoDB",
			Chart:       "bitnami/mongodb",
			Description: "NoSQL document database",
			RepoName:    "bitnami",
			RepoURL:     "https://charts.bitnami.com/bitnami",
		},
		{
			Name:        "MySQL",
			Chart:       "bitnami/mysql",
			Description: "Relational database",
			RepoName:    "bitnami",
			RepoURL:     "https://charts.bitnami.com/bitnami",
		},
		// Prometheus Community
		{
			Name:        "Prometheus",
			Chart:       "prometheus-community/prometheus",
			Description: "Monitoring and alerting toolkit",
			RepoName:    "prometheus-community",
			RepoURL:     "https://prometheus-community.github.io/helm-charts",
		},
		{
			Name:        "Grafana",
			Chart:       "grafana/grafana",
			Description: "Visualization and analytics platform",
			RepoName:    "grafana",
			RepoURL:     "https://grafana.github.io/helm-charts",
		},
		// Ingress
		{
			Name:        "NGINX Ingress",
			Chart:       "ingress-nginx/ingress-nginx",
			Description: "Ingress controller for Kubernetes",
			RepoName:    "ingress-nginx",
			RepoURL:     "https://kubernetes.github.io/ingress-nginx",
		},
		// Cert Manager
		{
			Name:        "Cert Manager",
			Chart:       "jetstack/cert-manager",
			Description: "X.509 certificate management",
			RepoName:    "jetstack",
			RepoURL:     "https://charts.jetstack.io",
		},
		// ArgoCD
		{
			Name:        "ArgoCD",
			Chart:       "argo/argo-cd",
			Description: "GitOps continuous delivery",
			RepoName:    "argo",
			RepoURL:     "https://argoproj.github.io/argo-helm",
		},
	}
}

// Init initializes the catalog
func (m *ChartCatalogModel) Init() tea.Cmd {
	return nil
}

// Update handles catalog messages
func (m *ChartCatalogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case quickStartMsg:
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

	return m, nil
}

// View renders the catalog
func (m *ChartCatalogModel) View() string {
	if m.installing {
		return ProcessingStyle.Render("⟳ Installing selected charts...")
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Chart Catalog - Select Charts to Install"))
	b.WriteString("\n\n")

	// Error message
	if m.err != nil {
		b.WriteString(ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err)))
		b.WriteString("\n\n")
	}

	// Chart list
	b.WriteString(m.renderChartList())

	// Help text
	b.WriteString("\n\n")
	selectedCount := len(m.selected)
	b.WriteString(HelpStyle.Render(fmt.Sprintf(
		"↑/↓: navigate • space: select/deselect (%d selected) • enter: install • esc: cancel",
		selectedCount,
	)))

	return b.String()
}

// renderChartList renders the list of charts
func (m *ChartCatalogModel) renderChartList() string {
	var b strings.Builder

	b.WriteString(HeaderStyle.Render(fmt.Sprintf("%-20s %-35s %s", "CHART", "NAME", "DESCRIPTION")))
	b.WriteString("\n")

	for i, chart := range m.charts {
		checkbox := "[ ]"
		if m.selected[i] {
			checkbox = "[✓]"
		}

		row := fmt.Sprintf("%s %-18s %-35s %s",
			checkbox,
			Truncate(chart.Name, 18),
			Truncate(chart.Chart, 35),
			Truncate(chart.Description, 40),
		)

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
func (m *ChartCatalogModel) MoveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveCursorDown moves the cursor down
func (m *ChartCatalogModel) MoveCursorDown() {
	if m.cursor < len(m.charts)-1 {
		m.cursor++
	}
}

// ToggleSelection toggles the selection of the current item
func (m *ChartCatalogModel) ToggleSelection() {
	if m.selected[m.cursor] {
		delete(m.selected, m.cursor)
	} else {
		m.selected[m.cursor] = true
	}
}

// GetSelectedCharts returns the selected charts
func (m *ChartCatalogModel) GetSelectedCharts() []ChartCatalogItem {
	var selected []ChartCatalogItem
	for i := range m.selected {
		if i < len(m.charts) {
			selected = append(selected, m.charts[i])
		}
	}
	return selected
}

// SetInstalling sets the installing state
func (m *ChartCatalogModel) SetInstalling(installing bool) {
	m.installing = installing
}

// Reset resets the catalog
func (m *ChartCatalogModel) Reset() {
	m.selected = make(map[int]bool)
	m.cursor = 0
	m.err = nil
}
