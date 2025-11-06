package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
)

// ValuesModel represents the values viewer
type ValuesModel struct {
	releaseName string
	values      map[string]interface{}
	yamlContent string
	viewport    viewport.Model
	loading     bool
	err         error
	ready       bool
	width       int
	height      int
}

// NewValuesModel creates a new values model
func NewValuesModel() *ValuesModel {
	vp := viewport.New(80, 20) // Default size
	return &ValuesModel{
		loading:  false,
		viewport: vp,
		width:    80,
		height:   20,
	}
}

// Init initializes the values viewer
func (m *ValuesModel) Init() tea.Cmd {
	return nil
}

// Update handles values viewer messages
func (m *ValuesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case valuesMsg:
		m.releaseName = msg.releaseName
		m.values = msg.values
		m.loading = false
		m.err = nil

		// Convert values to YAML
		yamlBytes, err := yaml.Marshal(m.values)
		if err != nil {
			m.err = fmt.Errorf("failed to convert values to YAML: %w", err)
			return m, nil
		}
		m.yamlContent = string(yamlBytes)

		// Update viewport content
		m.viewport.SetContent(m.yamlContent)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		m.ready = true
		return m, nil

	case errMsg:
		m.err = error(msg)
		m.loading = false
		return m, nil
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the values viewer
func (m *ValuesModel) View() string {
	if m.loading {
		return ProcessingStyle.Render("⟳ Loading values...")
	}

	if m.err != nil {
		return ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.releaseName == "" {
		return InfoMessageStyle.Render("No release selected")
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Values: %s", m.releaseName)))
	b.WriteString("\n\n")

	// Viewport with YAML content
	b.WriteString(CardStyle.Render(m.viewport.View()))

	// Help text
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("↑/↓: scroll • esc: back to releases • q: quit"))

	return b.String()
}

// SetLoading sets the loading state
func (m *ValuesModel) SetLoading(loading bool) {
	m.loading = loading
}
