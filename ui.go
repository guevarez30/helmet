package main

import (
	"os"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/pidanou/helm-tui/helpers"
	"github.com/pidanou/helm-tui/hub"
	"github.com/pidanou/helm-tui/plugins"
	"github.com/pidanou/helm-tui/releases"
	"github.com/pidanou/helm-tui/repositories"
	"github.com/pidanou/helm-tui/styles"
	"github.com/pidanou/helm-tui/types"
)

type tabIndex uint

var tabLabels = []string{"Releases", "Repositories", "Hub", "Plugins"}

const (
	releasesTab tabIndex = iota
	repositoriesTab
	hubTab
	pluginsTab
)

type viewState struct {
	tab         tabIndex
	breadcrumbs []string
}

var commandAliases = map[string]tabIndex{
	"releases":     releasesTab,
	"rel":          releasesTab,
	"repos":        repositoriesTab,
	"repo":         repositoriesTab,
	"repositories": repositoriesTab,
	"hub":          hubTab,
	"plugins":      pluginsTab,
	"plug":         pluginsTab,
}

var commandNames = []string{"releases", "repos", "hub", "plugins", "quit"}

type mainModel struct {
	state            tabIndex
	width            int
	height           int
	tabs             []string
	tabContent       []tea.Model
	loaded           bool
	commandBar       textinput.Model
	commandBarActive bool
	breadcrumbs      []string
	statusText       string
	statusFlash      string
	history          []viewState
	historyIdx       int
}

func newModel(tabs []string) mainModel {
	cb := textinput.New()
	cb.Placeholder = "type a command..."
	cb.CharLimit = 64
	cb.SetSuggestions(commandNames)
	cb.ShowSuggestions = true

	m := mainModel{
		state:       releasesTab,
		tabs:        tabs,
		tabContent:  make([]tea.Model, len(tabs)),
		loaded:      false,
		commandBar:  cb,
		breadcrumbs: []string{},
		history:     []viewState{},
		historyIdx:  -1,
	}
	m.tabContent[releasesTab], _ = releases.InitModel()
	m.tabContent[repositoriesTab], _ = repositories.InitModel()
	m.tabContent[hubTab] = hub.InitModel()
	m.tabContent[pluginsTab] = plugins.InitModel()
	return m
}

func (m mainModel) Init() tea.Cmd {
	var cmds = []tea.Cmd{createWorkingDir, textinput.Blink}
	for _, i := range m.tabContent {
		cmds = append(cmds, i.Init())
	}
	return tea.Batch(cmds...)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case types.InitAppMsg:
		if msg.Err != nil {
			return m, tea.Quit
		}
		m.loaded = true
	case types.EditorFinishedMsg:
		switch m.state {
		case releasesTab:
			m.tabContent[releasesTab], cmd = m.tabContent[releasesTab].Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		case repositoriesTab:
			m.tabContent[repositoriesTab], cmd = m.tabContent[repositoriesTab].Update(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}
	case types.BreadcrumbMsg:
		m.breadcrumbs = msg.Crumbs
	case types.StatusMsg:
		m.statusFlash = msg.Text
		return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
			return types.ClearFlashMsg{}
		})
	case types.ClearFlashMsg:
		m.statusFlash = ""
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		chromeHeight := 2 // header + status bar
		adjustedMsg := tea.WindowSizeMsg{Width: m.width, Height: msg.Height - chromeHeight}
		for i := range m.tabContent {
			m.tabContent[i], cmd = m.tabContent[i].Update(adjustedMsg)
			cmds = append(cmds, cmd)
		}
		m.commandBar.Width = m.width - 3
		return m, tea.Batch(cmds...)
	case tea.KeyMsg:
		// Command bar intercepts all keys when active
		if m.commandBarActive {
			switch msg.String() {
			case "enter":
				m.commandBarActive = false
				m.commandBar.Blur()
				cmd = m.executeCommand(m.commandBar.Value())
				m.commandBar.SetValue("")
				if cmd != nil {
					return m, cmd
				}
				return m, nil
			case "esc":
				m.commandBarActive = false
				m.commandBar.Blur()
				m.commandBar.SetValue("")
				return m, nil
			case "ctrl+c":
				return m, tea.Quit
			}
			m.commandBar, cmd = m.commandBar.Update(msg)
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case ":":
			m.commandBarActive = true
			m.commandBar.Focus()
			m.commandBar.SetValue("")
			return m, textinput.Blink
		case "]":
			// History forward
			if m.historyIdx < len(m.history)-1 {
				m.historyIdx++
				state := m.history[m.historyIdx]
				m.state = state.tab
				m.breadcrumbs = state.breadcrumbs
			}
			return m, nil
		case "[":
			// History back
			if m.historyIdx > 0 {
				m.historyIdx--
				state := m.history[m.historyIdx]
				m.state = state.tab
				m.breadcrumbs = state.breadcrumbs
			}
			return m, nil
		case "tab":
			// Cycle forward through tabs
			m.pushHistory()
			if m.state == pluginsTab {
				m.state = releasesTab
			} else {
				m.state++
			}
			m.breadcrumbs = []string{}
			m.tabContent[m.state], cmd = m.tabContent[m.state].Update(types.ResetViewMsg{})
			return m, cmd
		case "shift+tab":
			// Cycle backward through tabs
			m.pushHistory()
			if m.state == releasesTab {
				m.state = pluginsTab
			} else {
				m.state--
			}
			m.breadcrumbs = []string{}
			m.tabContent[m.state], cmd = m.tabContent[m.state].Update(types.ResetViewMsg{})
			return m, cmd
		}

		// Forward key to active tab only
		m.tabContent[m.state], cmd = m.tabContent[m.state].Update(msg)
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}
	// Non-key messages go to all tabs
	for i := range m.tabContent {
		m.tabContent[i], cmd = m.tabContent[i].Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *mainModel) executeCommand(input string) tea.Cmd {
	input = strings.TrimSpace(strings.ToLower(input))
	if input == "" {
		return nil
	}
	if input == "q" || input == "quit" {
		return tea.Quit
	}
	if target, ok := commandAliases[input]; ok {
		m.pushHistory()
		m.state = target
		m.breadcrumbs = []string{}
		var cmd tea.Cmd
		m.tabContent[m.state], cmd = m.tabContent[m.state].Update(types.ResetViewMsg{})
		return cmd
	}
	return nil
}

func (m *mainModel) pushHistory() {
	state := viewState{tab: m.state, breadcrumbs: append([]string{}, m.breadcrumbs...)}
	// Trim forward history if we navigated back then went elsewhere
	if m.historyIdx >= 0 && m.historyIdx < len(m.history)-1 {
		m.history = m.history[:m.historyIdx+1]
	}
	m.history = append(m.history, state)
	m.historyIdx = len(m.history) - 1
	// Cap at 50
	if len(m.history) > 50 {
		m.history = m.history[len(m.history)-50:]
		m.historyIdx = len(m.history) - 1
	}
}

func (m mainModel) View() string {
	if !m.loaded || len(m.tabContent) == 0 {
		return "loading..."
	}
	header := m.renderHeader()
	content := m.tabContent[m.state].View()
	statusBar := m.renderStatusBar()
	return header + "\n" + content + "\n" + statusBar
}

func (m mainModel) renderHeader() string {
	// Left side: breadcrumb path
	crumbs := []string{m.tabs[m.state]}
	crumbs = append(crumbs, m.breadcrumbs...)
	pathStr := strings.Join(crumbs, " > ")
	pathStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.HighlightColor)
	left := pathStyle.Render(pathStr)

	// Right side: tab indicators
	var tabIndicators []string
	for i, t := range m.tabs {
		if i == int(m.state) {
			tabIndicators = append(tabIndicators, lipgloss.NewStyle().
				Bold(true).
				Foreground(styles.HighlightColor).
				Render(t))
		} else {
			tabIndicators = append(tabIndicators, lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Render(t))
		}
	}
	right := strings.Join(tabIndicators, " | ")

	gap := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func (m mainModel) renderStatusBar() string {
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	var left string
	if m.commandBarActive {
		left = ":" + m.commandBar.View()
	} else if m.statusFlash != "" {
		left = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(m.statusFlash)
	}

	right := statusStyle.Render(m.statusText)
	gap := max(0, m.width-lipgloss.Width(left)-lipgloss.Width(right))
	return left + strings.Repeat(" ", gap) + right
}

func createWorkingDir() tea.Msg {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return types.InitAppMsg{Err: err}
	}
	workingDir := path.Join(homeDir, ".helm-tui")
	err = os.MkdirAll(workingDir, 0755)
	if err != nil {
		return types.InitAppMsg{Err: err}
	}
	helpers.UserDir = workingDir
	return types.InitAppMsg{Err: nil}
}
