package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// HelpModel represents the help view
type HelpModel struct {
	viewport viewport.Model
	ready    bool
	width    int
	height   int
}

// NewHelpModel creates a new help model
func NewHelpModel() *HelpModel {
	vp := viewport.New(80, 20) // Default size
	return &HelpModel{
		viewport: vp,
		width:    80,
		height:   20,
	}
}

// Init initializes the help view
func (m *HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles help view messages
func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 10
		m.ready = true
		// Set content after viewport is sized
		m.viewport.SetContent(m.getHelpContent())
		return m, nil
	}

	// Update viewport for scrolling
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the help view
func (m *HelpModel) View() string {
	if !m.ready {
		return InfoMessageStyle.Render("Loading help...")
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render("🪖 Helmet - User Guide"))
	b.WriteString("\n\n")

	// Viewport with scrollable content
	b.WriteString(m.viewport.View())

	// Footer help with scroll instructions
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓ or j/k: scroll • Tab: switch views • q: quit"))

	return b.String()
}

// getHelpContent generates the help content string
func (m *HelpModel) getHelpContent() string {
	var b strings.Builder

	// Overview section
	b.WriteString(CardTitleStyle.Render("Overview"))
	b.WriteString("\n")
	b.WriteString("Helmet is a Terminal User Interface (TUI) for managing Helm releases and repositories\n")
	b.WriteString("in Kubernetes clusters. Navigate with Tab, control with keyboard shortcuts.\n")
	b.WriteString("\n")

	// Help Tab section
	b.WriteString(SeparatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n")
	b.WriteString(CardTitleStyle.Render("📖 Help Tab (Current)"))
	b.WriteString("\n")
	b.WriteString("This tab provides documentation for all features and keyboard shortcuts.\n")
	b.WriteString("  " + InfoMessageStyle.Render("Tab") + " - Cycle through tabs\n")
	b.WriteString("  " + InfoMessageStyle.Render("↑/↓") + " or " + InfoMessageStyle.Render("j/k") + " - Scroll through help content\n")
	b.WriteString("\n")

	// Releases Tab section
	b.WriteString(SeparatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n")
	b.WriteString(CardTitleStyle.Render("🚀 Releases Tab"))
	b.WriteString("\n")
	b.WriteString("View, manage, and monitor all Helm releases in your current namespace.\n")
	b.WriteString("\n")
	b.WriteString(SuccessMessageStyle.Render("Actions:") + "\n")
	b.WriteString("  " + InfoMessageStyle.Render("↑/↓") + " or " + InfoMessageStyle.Render("j/k") + " - Navigate through releases\n")
	b.WriteString("  " + InfoMessageStyle.Render("i") + "           - Install a new chart\n")
	b.WriteString("  " + InfoMessageStyle.Render("u") + "           - Upgrade selected release to a new version\n")
	b.WriteString("  " + InfoMessageStyle.Render("d") + "           - Delete/uninstall selected release\n")
	b.WriteString("  " + InfoMessageStyle.Render("v") + "           - View release values (YAML configuration)\n")
	b.WriteString("  " + InfoMessageStyle.Render("p") + "           - View pods associated with release\n")
	b.WriteString("  " + InfoMessageStyle.Render("Ctrl+R") + "      - Refresh release list\n")
	b.WriteString("\n")
	b.WriteString(SuccessMessageStyle.Render("Features:") + "\n")
	b.WriteString("  • View release status (" + DeployedStyle.Render("deployed") + ", " + FailedStyle.Render("failed") + ", " + PendingStyle.Render("pending") + ")\n")
	b.WriteString("  • See chart versions and update times\n")
	b.WriteString("  • Quick access to pod logs and values\n")
	b.WriteString("\n")

	// Repositories Tab section
	b.WriteString(SeparatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n")
	b.WriteString(CardTitleStyle.Render("📚 Repositories Tab"))
	b.WriteString("\n")
	b.WriteString("Manage Helm chart repositories (HTTP, local, OCI).\n")
	b.WriteString("\n")
	b.WriteString(SuccessMessageStyle.Render("Actions:") + "\n")
	b.WriteString("  " + InfoMessageStyle.Render("↑/↓") + " or " + InfoMessageStyle.Render("j/k") + " - Navigate through repositories\n")
	b.WriteString("  " + InfoMessageStyle.Render("a") + "           - Add a new repository\n")
	b.WriteString("  " + InfoMessageStyle.Render("r") + "           - Remove selected repository\n")
	b.WriteString("  " + InfoMessageStyle.Render("U") + "           - Update all repositories (fetch latest charts)\n")
	b.WriteString("  " + InfoMessageStyle.Render("Ctrl+R") + "      - Refresh repository list\n")
	b.WriteString("\n")
	b.WriteString(SuccessMessageStyle.Render("Features:") + "\n")
	b.WriteString("  • View all configured Helm repositories\n")
	b.WriteString("  • Add custom chart repositories (HTTP, file://, oci://)\n")
	b.WriteString("  • Keep repositories up to date\n")
	b.WriteString("\n")

	// Global Shortcuts section
	b.WriteString(SeparatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n")
	b.WriteString(CardTitleStyle.Render("⌨️  Global Shortcuts"))
	b.WriteString("\n")
	b.WriteString("  " + InfoMessageStyle.Render("Tab") + "    - Switch between tabs\n")
	b.WriteString("  " + InfoMessageStyle.Render("q") + "      - Quit application (with confirmation)\n")
	b.WriteString("  " + InfoMessageStyle.Render("Ctrl+C") + " - Force quit\n")
	b.WriteString("  " + InfoMessageStyle.Render("Esc") + "    - Close modal views and return to previous view\n")
	b.WriteString("\n")

	// Context Information section
	b.WriteString(SeparatorStyle.Render("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"))
	b.WriteString("\n")
	b.WriteString(CardTitleStyle.Render("💡 Tips"))
	b.WriteString("\n")
	b.WriteString("  • Check the " + NamespaceStyle.Render("status bar") + " at the bottom for current context and namespace\n")
	b.WriteString("  • Install operations target the namespace shown in the status bar\n")
	b.WriteString("  • Use " + InfoMessageStyle.Render("v") + " on any release to inspect its configuration\n")
	b.WriteString("  • Use " + InfoMessageStyle.Render("p") + " to quickly access pod logs for debugging\n")
	b.WriteString("  • Repository updates are recommended before installing new charts\n")

	return b.String()
}
