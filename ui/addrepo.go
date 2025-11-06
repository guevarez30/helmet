package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// AddRepoModel represents the add repository form
type AddRepoModel struct {
	repoName  textinput.Model
	repoURL   textinput.Model
	focusIndex int
	inputs     []textinput.Model
	err        error
	adding     bool
}

// NewAddRepoModel creates a new add repository model
func NewAddRepoModel() *AddRepoModel {
	m := &AddRepoModel{
		repoName:   textinput.New(),
		repoURL:    textinput.New(),
		focusIndex: 0,
	}

	m.repoName.Placeholder = "my-repo"
	m.repoName.Focus()
	m.repoName.CharLimit = 50
	m.repoName.Width = 50
	m.repoName.Prompt = "Repository Name: "

	m.repoURL.Placeholder = "https://example.com/charts or file:///path/to/charts"
	m.repoURL.CharLimit = 200
	m.repoURL.Width = 50
	m.repoURL.Prompt = "Repository URL:  "

	m.inputs = []textinput.Model{m.repoName, m.repoURL}

	return m
}

// Init initializes the add repository form
func (m *AddRepoModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles add repository form messages
func (m *AddRepoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case repoActionMsg:
		m.adding = false
		if !msg.success {
			m.err = fmt.Errorf("%s", msg.message)
		}
		return m, nil

	case errMsg:
		m.err = error(msg)
		m.adding = false
		return m, nil
	}

	// Handle input updates
	var cmd tea.Cmd
	m.inputs[m.focusIndex], cmd = m.inputs[m.focusIndex].Update(msg)

	// Update the actual fields
	m.repoName = m.inputs[0]
	m.repoURL = m.inputs[1]

	return m, cmd
}

// View renders the add repository form
func (m *AddRepoModel) View() string {
	if m.adding {
		return ProcessingStyle.Render("⟳ Adding repository...")
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("Add Helm Repository"))
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
	b.WriteString(HelpStyle.Render("tab/shift+tab: switch field • enter: add • esc: cancel"))

	return b.String()
}

// renderForm renders the form fields
func (m *AddRepoModel) renderForm() string {
	var b strings.Builder

	b.WriteString("Add a new Helm chart repository:\n\n")

	for i, input := range m.inputs {
		b.WriteString(input.View())
		if i < len(m.inputs)-1 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n\n")
	b.WriteString("Examples:\n")
	b.WriteString("  HTTP/HTTPS:  https://charts.bitnami.com/bitnami\n")
	b.WriteString("  Local:       file:///Users/you/my-charts\n")
	b.WriteString("  OCI:         oci://registry.example.com/charts\n")

	return b.String()
}

// MoveFocusNext moves focus to the next input
func (m *AddRepoModel) MoveFocusNext() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex = (m.focusIndex + 1) % len(m.inputs)
	m.inputs[m.focusIndex].Focus()
}

// MoveFocusPrev moves focus to the previous input
func (m *AddRepoModel) MoveFocusPrev() {
	m.inputs[m.focusIndex].Blur()
	m.focusIndex--
	if m.focusIndex < 0 {
		m.focusIndex = len(m.inputs) - 1
	}
	m.inputs[m.focusIndex].Focus()
}

// GetRepoName returns the repository name
func (m *AddRepoModel) GetRepoName() string {
	name := m.repoName.Value()
	if name == "" {
		return m.repoName.Placeholder
	}
	return name
}

// GetRepoURL returns the repository URL
func (m *AddRepoModel) GetRepoURL() string {
	return m.repoURL.Value()
}

// SetAdding sets the adding state
func (m *AddRepoModel) SetAdding(adding bool) {
	m.adding = adding
}

// Reset resets the form
func (m *AddRepoModel) Reset() {
	m.repoName.SetValue("")
	m.repoURL.SetValue("")
	m.err = nil
	m.focusIndex = 0
	m.inputs[0].Focus()
}
