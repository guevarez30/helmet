package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/guevarez30/helmet/helm"
	"github.com/guevarez30/helmet/kubernetes"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
)

// View represents different application views
type View int

const (
	HelpView View = iota
	ReleasesView
	RepositoriesView
	ValuesView
	InstallView
	CatalogView
	AddRepoView
	UpgradeView
	PodsView
	LogsView
)

// Model represents the main application model
type Model struct {
	helmClient  *helm.Client
	k8sContext  *kubernetes.ContextManager
	currentView View
	width       int
	height      int
	keys        KeyMap

	// Sub-models
	help          *HelpModel
	releases      *ReleasesModel
	repositories  *RepositoriesModel
	values        *ValuesModel
	install       *InstallModel
	catalog       *ChartCatalogModel
	addRepo       *AddRepoModel
	upgrade       *UpgradeModel
	pods          *PodsModel
	logs          *LogsModel

	// State
	previousView         View // For returning from modal views
	showQuitConfirmation bool // Show quit confirmation dialog
	err                  error
}

// Message types
type releasesMsg []*release.Release
type releaseActionMsg struct {
	success bool
	message string
}
type repositoriesMsg []*repo.Entry
type repoActionMsg struct {
	success bool
	message string
}
type valuesMsg struct {
	releaseName string
	values      map[string]interface{}
}
type installActionMsg struct {
	success bool
	message string
}
type quickStartMsg struct {
	success bool
	message string
}
type upgradeVersionsMsg struct {
	versions []ChartVersion
}
type upgradeActionMsg struct {
	success bool
	message string
}
type podsMsg struct {
	releaseName string
	pods        []corev1.Pod
}
type logsMsg struct {
	podName string
	logs    string
}
type clearStatusMsg struct{}
type errMsg error

// NewModel creates a new application model
func NewModel() (*Model, error) {
	// Initialize Kubernetes context manager
	k8sContext, err := kubernetes.NewContextManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize kubernetes context: %w", err)
	}

	// Initialize Helm client with current namespace
	helmClient, err := helm.NewClient(k8sContext.GetCurrentNamespace())
	if err != nil {
		return nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}

	m := &Model{
		helmClient:   helmClient,
		k8sContext:   k8sContext,
		currentView:  HelpView,
		keys:         DefaultKeyMap(),
		help:         NewHelpModel(),
		releases:     NewReleasesModel(),
		repositories: NewRepositoriesModel(),
		values:       NewValuesModel(),
		install:      NewInstallModel(),
		catalog:      NewChartCatalogModel(),
		addRepo:      NewAddRepoModel(),
		upgrade:      NewUpgradeModel(),
		pods:         NewPodsModel(),
		logs:         NewLogsModel(),
	}

	return m, nil
}

// Init initializes the application
func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		m.refreshReleases(),
		m.refreshRepositories(),
	)
}

// Update handles messages
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Forward window size to specific views
		if m.currentView == HelpView {
			var cmd tea.Cmd
			_, cmd = m.help.Update(msg)
			return m, cmd
		}
		if m.currentView == ValuesView {
			var cmd tea.Cmd
			_, cmd = m.values.Update(msg)
			return m, cmd
		}
		if m.currentView == LogsView {
			var cmd tea.Cmd
			_, cmd = m.logs.Update(msg)
			return m, cmd
		}
		return m, nil

	case installActionMsg:
		// Handle install completion
		m.install.SetInstalling(false)
		if msg.success {
			// Return to previous view and refresh releases
			m.currentView = m.previousView
			m.install.Reset()
			return m, m.refreshReleases()
		} else {
			// Stay in install view to show error
			var cmd tea.Cmd
			_, cmd = m.install.Update(msg)
			return m, cmd
		}
		return m, nil

	case repoActionMsg:
		// Handle repo action completion (could be from add or remove)
		m.addRepo.SetAdding(false)
		if msg.success {
			// If we're in AddRepoView, return to repositories
			if m.currentView == AddRepoView {
				m.currentView = m.previousView
				m.addRepo.Reset()
			}
			// Refresh repositories regardless
			return m, m.refreshRepositories()
		} else {
			// Stay in current view to show error
			if m.currentView == AddRepoView {
				var cmd tea.Cmd
				_, cmd = m.addRepo.Update(msg)
				return m, cmd
			}
		}
		return m, nil

	case quickStartMsg:
		// Handle quick start completion
		m.catalog.SetInstalling(false)
		if msg.success {
			// Return to previous view and refresh
			m.currentView = m.previousView
			return m, tea.Batch(
				m.refreshReleases(),
				m.refreshRepositories(),
			)
		} else {
			// Stay in catalog view to show error
			var cmd tea.Cmd
			_, cmd = m.catalog.Update(msg)
			return m, cmd
		}
		return m, nil

	case upgradeVersionsMsg:
		// Forward to upgrade view
		var cmd tea.Cmd
		_, cmd = m.upgrade.Update(msg)
		return m, cmd

	case upgradeActionMsg:
		// Handle upgrade completion
		m.upgrade.SetUpgrading(false)
		if msg.success {
			// Return to releases view and refresh
			m.currentView = m.previousView
			m.upgrade.Reset()
			return m, m.refreshReleases()
		} else {
			// Stay in upgrade view to show error
			var cmd tea.Cmd
			_, cmd = m.upgrade.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case releasesMsg:
		m.releases.SetReleases(msg)
		return m, nil

	case repositoriesMsg:
		m.repositories.SetRepositories(msg)
		return m, nil

	case valuesMsg:
		// Switch to values view when values are loaded
		m.previousView = m.currentView
		m.currentView = ValuesView
		var cmd tea.Cmd
		_, cmd = m.values.Update(msg)
		// Send window size to initialize viewport
		_, sizeCmd := m.values.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		return m, tea.Batch(cmd, sizeCmd)

	case podsMsg:
		// Switch to pods view when pods are loaded
		m.previousView = m.currentView
		m.currentView = PodsView
		var cmd tea.Cmd
		_, cmd = m.pods.Update(msg)
		// Send window size to initialize viewport if needed
		_, sizeCmd := m.pods.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		return m, tea.Batch(cmd, sizeCmd)

	case logsMsg:
		// Switch to logs view when logs are loaded
		// previousView is already set from when we entered PodsView (should be ReleasesView)
		m.currentView = LogsView
		var cmd tea.Cmd
		_, cmd = m.logs.Update(msg)
		// Send window size to initialize viewport
		_, sizeCmd := m.logs.Update(tea.WindowSizeMsg{Width: m.width, Height: m.height})
		return m, tea.Batch(cmd, sizeCmd)

	case errMsg:
		m.err = error(msg)
		return m, nil
	}

	// Route messages to active view
	switch m.currentView {
	case HelpView:
		var cmd tea.Cmd
		_, cmd = m.help.Update(msg)
		return m, cmd
	case ReleasesView:
		var cmd tea.Cmd
		_, cmd = m.releases.Update(msg)
		return m, cmd
	case RepositoriesView:
		var cmd tea.Cmd
		_, cmd = m.repositories.Update(msg)
		return m, cmd
	case ValuesView:
		var cmd tea.Cmd
		_, cmd = m.values.Update(msg)
		return m, cmd
	case InstallView:
		var cmd tea.Cmd
		_, cmd = m.install.Update(msg)
		return m, cmd
	case CatalogView:
		var cmd tea.Cmd
		_, cmd = m.catalog.Update(msg)
		return m, cmd
	case AddRepoView:
		var cmd tea.Cmd
		_, cmd = m.addRepo.Update(msg)
		return m, cmd
	case UpgradeView:
		var cmd tea.Cmd
		_, cmd = m.upgrade.Update(msg)
		return m, cmd
	case PodsView:
		var cmd tea.Cmd
		_, cmd = m.pods.Update(msg)
		return m, cmd
	case LogsView:
		var cmd tea.Cmd
		_, cmd = m.logs.Update(msg)
		return m, cmd
	}

	return m, nil
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle quit confirmation dialog
	if m.showQuitConfirmation {
		switch msg.String() {
		case "y", "Y", "enter":
			return m, tea.Quit
		case "n", "N", "esc":
			m.showQuitConfirmation = false
			return m, nil
		}
		return m, nil
	}

	// Check for quit key (q) in all views
	if msg.String() == "q" {
		m.showQuitConfirmation = true
		return m, nil
	}

	// Install view gets priority for most keypresses (except quit and escape)
	if m.currentView == InstallView {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentView = m.previousView
			m.install.Reset()
			return m, nil
		case "tab":
			m.install.MoveFocusNext()
			return m, nil
		case "shift+tab":
			m.install.MoveFocusPrev()
			return m, nil
		case "enter":
			return m, m.installChart()
		default:
			// Forward all other keys to the install view for text input
			var cmd tea.Cmd
			_, cmd = m.install.Update(msg)
			return m, cmd
		}
	}

	// Catalog view keypresses
	if m.currentView == CatalogView {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentView = m.previousView
			m.catalog.Reset()
			return m, nil
		case "up", "k":
			m.catalog.MoveCursorUp()
			return m, nil
		case "down", "j":
			m.catalog.MoveCursorDown()
			return m, nil
		case " ":
			m.catalog.ToggleSelection()
			return m, nil
		case "enter":
			return m, m.installSelectedCharts()
		}
		return m, nil
	}

	// Add repository view keypresses
	if m.currentView == AddRepoView {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentView = m.previousView
			m.addRepo.Reset()
			return m, nil
		case "tab":
			m.addRepo.MoveFocusNext()
			return m, nil
		case "shift+tab":
			m.addRepo.MoveFocusPrev()
			return m, nil
		case "enter":
			return m, m.addRepository()
		default:
			// Forward all other keys for text input
			var cmd tea.Cmd
			_, cmd = m.addRepo.Update(msg)
			return m, cmd
		}
	}

	// Upgrade view keypresses
	if m.currentView == UpgradeView {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentView = m.previousView
			m.upgrade.Reset()
			return m, nil
		case "up", "k":
			m.upgrade.MoveCursorUp()
			return m, nil
		case "down", "j":
			m.upgrade.MoveCursorDown()
			return m, nil
		case "enter":
			return m, m.upgradeRelease()
		}
		return m, nil
	}

	// Pods view keypresses
	if m.currentView == PodsView {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.currentView = m.previousView
			m.pods.Reset()
			return m, nil
		case "up", "k":
			m.pods.MoveCursorUp()
			return m, nil
		case "down", "j":
			m.pods.MoveCursorDown()
			return m, nil
		case "l":
			return m, m.viewPodLogs()
		}
		return m, nil
	}

	// Logs view keypresses
	if m.currentView == LogsView {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			// Always return to PodsView from LogsView
			m.currentView = PodsView
			m.logs.Reset()
			return m, nil
		default:
			// Forward other keys to viewport for scrolling
			var cmd tea.Cmd
			_, cmd = m.logs.Update(msg)
			return m, cmd
		}
	}

	switch {
	// Quit (only ctrl+c now, since 'q' is handled above with confirmation)
	case msg.String() == "ctrl+c":
		return m, tea.Quit

	// Escape - return from modal views (ValuesView only, others handled above)
	case msg.String() == "esc" && m.currentView == ValuesView:
		m.currentView = m.previousView
		return m, nil

	// Tab - switch views (but not in modal views)
	case msg.String() == "tab" && m.currentView != ValuesView && m.currentView != PodsView && m.currentView != LogsView:
		return m, m.switchView()

	// Refresh
	case msg.String() == "ctrl+r":
		switch m.currentView {
		case RepositoriesView:
			return m, m.refreshRepositories()
		default:
			return m, m.refreshReleases()
		}

	// View-specific navigation
	case msg.String() == "up", msg.String() == "k":
		switch m.currentView {
		case ReleasesView:
			m.releases.MoveCursorUp()
		case RepositoriesView:
			m.repositories.MoveCursorUp()
		}
		return m, nil

	case msg.String() == "down", msg.String() == "j":
		switch m.currentView {
		case ReleasesView:
			m.releases.MoveCursorDown()
		case RepositoriesView:
			m.repositories.MoveCursorDown()
		}
		return m, nil

	// Release operations (only in releases view)
	case msg.String() == "d" && m.currentView == ReleasesView:
		return m, m.deleteRelease()

	case msg.String() == "i" && m.currentView == ReleasesView:
		// Open install form
		m.previousView = m.currentView
		m.currentView = InstallView
		return m, m.install.Init()

	case msg.String() == "u" && m.currentView == ReleasesView:
		// Open upgrade view
		return m, m.openUpgradeView()

	case msg.String() == "H" && m.currentView == ReleasesView:
		return m, nil // TODO: implement history view

	case msg.String() == "v" && m.currentView == ReleasesView:
		return m, m.viewReleaseValues()

	case msg.String() == "p" && m.currentView == ReleasesView:
		return m, m.viewReleasePods()

	// Repository operations (only in repositories view)
	case msg.String() == "a" && m.currentView == RepositoriesView:
		// Open add repository form
		m.previousView = m.currentView
		m.currentView = AddRepoView
		return m, m.addRepo.Init()

	case msg.String() == "r" && m.currentView == RepositoriesView:
		return m, m.removeRepository()

	case msg.String() == "U" && m.currentView == RepositoriesView:
		return m, m.updateRepositories()
	}

	return m, nil
}

// View renders the UI
func (m *Model) View() string {
	var b strings.Builder

	// Render tabs
	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	// Render current view
	switch m.currentView {
	case HelpView:
		b.WriteString(m.help.View())
	case ReleasesView:
		b.WriteString(m.releases.View())
	case RepositoriesView:
		b.WriteString(m.repositories.View())
	case ValuesView:
		b.WriteString(m.values.View())
	case InstallView:
		b.WriteString(m.install.View())
	case CatalogView:
		b.WriteString(m.catalog.View())
	case AddRepoView:
		b.WriteString(m.addRepo.View())
	case UpgradeView:
		b.WriteString(m.upgrade.View())
	case PodsView:
		b.WriteString(m.pods.View())
	case LogsView:
		b.WriteString(m.logs.View())
	default:
		b.WriteString("Unknown view")
	}

	b.WriteString("\n")

	// Render status bar
	b.WriteString(m.renderStatusBar())

	// Overlay quit confirmation dialog if active
	if m.showQuitConfirmation {
		return m.overlayQuitConfirmation(b.String())
	}

	return b.String()
}

// overlayQuitConfirmation overlays a quit confirmation dialog on the current view
func (m *Model) overlayQuitConfirmation(baseView string) string {
	var b strings.Builder

	// Get base view lines
	lines := strings.Split(baseView, "\n")

	// Calculate dialog position (centered)
	dialogWidth := 52
	dialogHeight := 7

	// Create the dialog box with just text
	dialog := []string{
		"┌──────────────────────────────────────────────────┐",
		"│                                                  │",
		"│          Are you sure you want to quit?          │",
		"│                                                  │",
		"│         Press 'y' or Enter to confirm            │",
		"│         Press 'n' or Esc to cancel               │",
		"└──────────────────────────────────────────────────┘",
	}

	// Determine where to place the dialog (roughly centered)
	startLine := len(lines)/2 - dialogHeight/2
	if startLine < 0 {
		startLine = 0
	}

	// Overlay the dialog
	for i, line := range lines {
		if i >= startLine && i < startLine+dialogHeight {
			dialogLine := dialog[i-startLine]
			// Center the dialog horizontally
			padding := (m.width - dialogWidth) / 2
			if padding < 0 {
				padding = 0
			}
			b.WriteString(strings.Repeat(" ", padding))
			b.WriteString(DialogStyle.Render(dialogLine))
		} else {
			// Dim the background
			b.WriteString(DimStyle.Render(line))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// renderTabs renders the tab navigation
func (m *Model) renderTabs() string {
	tabs := []string{"Help", "Releases", "Repositories"}
	var renderedTabs []string

	for i, tab := range tabs {
		if View(i) == m.currentView {
			renderedTabs = append(renderedTabs, ActiveTabStyle.Render(tab))
		} else {
			renderedTabs = append(renderedTabs, InactiveTabStyle.Render(tab))
		}
	}

	return strings.Join(renderedTabs, TabSeparatorStyle.Render(" │ "))
}

// renderStatusBar renders the bottom status bar
func (m *Model) renderStatusBar() string {
	cluster, _, namespace := m.k8sContext.GetContextInfo()
	context := m.k8sContext.GetCurrentContext()

	status := fmt.Sprintf("Context: %s | Cluster: %s | Namespace: %s",
		ContextStyle.Render(context),
		InfoMessageStyle.Render(cluster),
		NamespaceStyle.Render(namespace))

	return StatusBarStyle.Render(status)
}

// switchView switches to the next view
func (m *Model) switchView() tea.Cmd {
	switch m.currentView {
	case HelpView:
		m.currentView = ReleasesView
	case ReleasesView:
		m.currentView = RepositoriesView
	case RepositoriesView:
		m.currentView = HelpView
	default:
		m.currentView = HelpView
	}
	return nil
}

// refreshReleases fetches the latest releases
func (m *Model) refreshReleases() tea.Cmd {
	return func() tea.Msg {
		releases, err := m.helmClient.ListReleases(true)
		if err != nil {
			return errMsg(err)
		}
		return releasesMsg(releases)
	}
}

// deleteRelease deletes the selected release
func (m *Model) deleteRelease() tea.Cmd {
	rel := m.releases.GetSelectedRelease()
	if rel == nil {
		return nil
	}

	m.releases.SetActionInProgress(true)

	return func() tea.Msg {
		_, err := m.helmClient.UninstallRelease(rel.Name)
		if err != nil {
			return releaseActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to delete release: %v", err),
			}
		}

		// Refresh releases after delete
		releases, err := m.helmClient.ListReleases(true)
		if err != nil {
			return errMsg(err)
		}

		// Send both action success and new releases
		return tea.Batch(
			func() tea.Msg {
				return releaseActionMsg{
					success: true,
					message: fmt.Sprintf("Release '%s' deleted successfully", rel.Name),
				}
			},
			func() tea.Msg {
				return releasesMsg(releases)
			},
		)()
	}
}

// refreshRepositories fetches the latest repositories
func (m *Model) refreshRepositories() tea.Cmd {
	return func() tea.Msg {
		repos, err := m.helmClient.ListRepositories()
		if err != nil {
			return errMsg(err)
		}
		return repositoriesMsg(repos)
	}
}

// removeRepository removes the selected repository
func (m *Model) removeRepository() tea.Cmd {
	repo := m.repositories.GetSelectedRepository()
	if repo == nil {
		return nil
	}

	m.repositories.SetActionInProgress(true)

	return func() tea.Msg {
		err := m.helmClient.RemoveRepository(repo.Name)
		if err != nil {
			return repoActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to remove repository: %v", err),
			}
		}

		// Refresh repositories after remove
		repos, err := m.helmClient.ListRepositories()
		if err != nil {
			return errMsg(err)
		}

		// Send both action success and new repositories
		return tea.Batch(
			func() tea.Msg {
				return repoActionMsg{
					success: true,
					message: fmt.Sprintf("Repository '%s' removed successfully", repo.Name),
				}
			},
			func() tea.Msg {
				return repositoriesMsg(repos)
			},
		)()
	}
}

// updateRepositories updates all repositories
func (m *Model) updateRepositories() tea.Cmd {
	m.repositories.SetActionInProgress(true)

	return func() tea.Msg {
		err := m.helmClient.UpdateRepositories(context.Background())
		if err != nil {
			return repoActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to update repositories: %v", err),
			}
		}

		// Refresh repositories after update
		repos, err := m.helmClient.ListRepositories()
		if err != nil {
			return errMsg(err)
		}

		// Send both action success and new repositories
		return tea.Batch(
			func() tea.Msg {
				return repoActionMsg{
					success: true,
					message: "All repositories updated successfully",
				}
			},
			func() tea.Msg {
				return repositoriesMsg(repos)
			},
		)()
	}
}

// viewReleaseValues displays the values for the selected release
func (m *Model) viewReleaseValues() tea.Cmd {
	rel := m.releases.GetSelectedRelease()
	if rel == nil {
		return nil
	}

	m.values.SetLoading(true)

	return func() tea.Msg {
		values, err := m.helmClient.GetReleaseValues(rel.Name)
		if err != nil {
			return errMsg(err)
		}

		return valuesMsg{
			releaseName: rel.Name,
			values:      values,
		}
	}
}

// viewReleasePods displays the pods for the selected release
func (m *Model) viewReleasePods() tea.Cmd {
	rel := m.releases.GetSelectedRelease()
	if rel == nil {
		return nil
	}

	m.pods.SetLoading(true)

	return func() tea.Msg {
		pods, err := m.helmClient.GetReleasePods(rel.Name)
		if err != nil {
			return errMsg(err)
		}

		return podsMsg{
			releaseName: rel.Name,
			pods:        pods,
		}
	}
}

// viewPodLogs displays the logs for the selected pod
func (m *Model) viewPodLogs() tea.Cmd {
	pod := m.pods.GetSelectedPod()
	if pod == nil {
		return nil
	}

	m.logs.SetLoading(true)

	return func() tea.Msg {
		logs, err := m.k8sContext.GetPodLogs(pod.Namespace, pod.Name, 100)
		if err != nil {
			return errMsg(err)
		}

		return logsMsg{
			podName: pod.Name,
			logs:    logs,
		}
	}
}

// installChart installs a chart from the install form
func (m *Model) installChart() tea.Cmd {
	releaseName := m.install.GetReleaseName()
	chartPath := m.install.GetChartPath()
	namespace := m.install.GetNamespace()

	if chartPath == "" {
		return func() tea.Msg {
			return installActionMsg{
				success: false,
				message: "Chart path is required",
			}
		}
	}

	m.install.SetInstalling(true)

	return func() tea.Msg {
		// Create new client with target namespace
		client, err := helm.NewClient(namespace)
		if err != nil {
			return installActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to create Helm client: %v", err),
			}
		}

		// Install chart with empty values
		_, err = client.InstallChart(releaseName, chartPath, nil)
		if err != nil {
			return installActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to install chart: %v", err),
			}
		}

		return installActionMsg{
			success: true,
			message: fmt.Sprintf("Chart '%s' installed successfully as '%s'", chartPath, releaseName),
		}
	}
}

// addRepository adds a new Helm repository
func (m *Model) addRepository() tea.Cmd {
	repoName := m.addRepo.GetRepoName()
	repoURL := m.addRepo.GetRepoURL()

	if repoURL == "" {
		return func() tea.Msg {
			return repoActionMsg{
				success: false,
				message: "Repository URL is required",
			}
		}
	}

	m.addRepo.SetAdding(true)

	return func() tea.Msg {
		err := m.helmClient.AddRepository(repoName, repoURL)
		if err != nil {
			return repoActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to add repository: %v", err),
			}
		}

		// Update repositories after adding
		err = m.helmClient.UpdateRepositories(context.Background())
		if err != nil {
			return repoActionMsg{
				success: false,
				message: fmt.Sprintf("Repository added but failed to update: %v", err),
			}
		}

		// Refresh repositories list
		repos, err := m.helmClient.ListRepositories()
		if err != nil {
			return errMsg(err)
		}

		// Send both action success and new repositories
		return tea.Batch(
			func() tea.Msg {
				return repoActionMsg{
					success: true,
					message: fmt.Sprintf("Repository '%s' added successfully", repoName),
				}
			},
			func() tea.Msg {
				return repositoriesMsg(repos)
			},
		)()
	}
}

// installSelectedCharts installs the charts selected from the catalog
func (m *Model) installSelectedCharts() tea.Cmd {
	selectedCharts := m.catalog.GetSelectedCharts()

	if len(selectedCharts) == 0 {
		return func() tea.Msg {
			return quickStartMsg{
				success: false,
				message: "No charts selected",
			}
		}
	}

	m.catalog.SetInstalling(true)

	return func() tea.Msg {
		// Get existing repos
		existingRepos, err := m.helmClient.ListRepositories()
		if err != nil {
			return quickStartMsg{
				success: false,
				message: fmt.Sprintf("Failed to list repositories: %v", err),
			}
		}

		// Build map of existing repos
		repoMap := make(map[string]bool)
		for _, repo := range existingRepos {
			repoMap[repo.Name] = true
		}

		// Add any missing repositories
		addedRepos := make(map[string]bool)
		for _, chart := range selectedCharts {
			if !repoMap[chart.RepoName] && !addedRepos[chart.RepoName] {
				err := m.helmClient.AddRepository(chart.RepoName, chart.RepoURL)
				if err != nil {
					return quickStartMsg{
						success: false,
						message: fmt.Sprintf("Failed to add repository %s: %v", chart.RepoName, err),
					}
				}
				addedRepos[chart.RepoName] = true
			}
		}

		// Update repositories if we added any
		if len(addedRepos) > 0 {
			err := m.helmClient.UpdateRepositories(context.Background())
			if err != nil {
				return quickStartMsg{
					success: false,
					message: fmt.Sprintf("Failed to update repositories: %v", err),
				}
			}
		}

		// Install selected charts
		installedCount := 0
		for _, chart := range selectedCharts {
			// Generate release name from chart name
			releaseName := "my-" + strings.ToLower(strings.ReplaceAll(chart.Name, " ", "-"))

			// Simple default values
			vals := map[string]interface{}{}
			if strings.Contains(chart.Chart, "redis") {
				vals["auth"] = map[string]interface{}{"enabled": false}
			}
			if strings.Contains(chart.Chart, "nginx") {
				vals["service"] = map[string]interface{}{"type": "ClusterIP"}
			}

			_, err := m.helmClient.InstallChart(releaseName, chart.Chart, vals)
			if err != nil {
				// Continue if chart already exists
				continue
			}
			installedCount++
		}

		m.catalog.Reset()

		return quickStartMsg{
			success: true,
			message: fmt.Sprintf("Successfully installed %d charts", installedCount),
		}
	}
}

// openUpgradeView opens the upgrade view for the selected release
func (m *Model) openUpgradeView() tea.Cmd {
	rel := m.releases.GetSelectedRelease()
	if rel == nil {
		return nil
	}

	// Set up upgrade view
	m.upgrade.SetRelease(rel.Name, rel.Chart.Metadata.Name, rel.Chart.Metadata.Version)
	m.previousView = m.currentView
	m.currentView = UpgradeView

	// Fetch available versions
	return m.fetchChartVersions(rel.Chart.Metadata.Name)
}

// fetchChartVersions fetches available versions for a chart
func (m *Model) fetchChartVersions(chartName string) tea.Cmd {
	return func() tea.Msg {
		versions, err := m.helmClient.GetChartVersions(chartName)
		if err != nil {
			return errMsg(err)
		}

		// Convert to UI chart versions
		var uiVersions []ChartVersion
		for _, ver := range versions {
			uiVersions = append(uiVersions, ChartVersion{
				Version:     ver.Version,
				AppVersion:  ver.AppVersion,
				Description: ver.Description,
			})
		}

		return upgradeVersionsMsg{versions: uiVersions}
	}
}

// upgradeRelease upgrades the release to the selected version
func (m *Model) upgradeRelease() tea.Cmd {
	selectedVer := m.upgrade.GetSelectedVersion()
	if selectedVer == nil {
		return nil
	}

	m.upgrade.SetUpgrading(true)

	return func() tea.Msg {
		// Get release info
		relName := m.upgrade.releaseName
		chartName := m.upgrade.chartName

		// Build chart reference with version
		chartRef := fmt.Sprintf("%s --version %s", chartName, selectedVer.Version)

		// Upgrade the release (using existing values)
		_, err := m.helmClient.UpgradeRelease(relName, chartRef, nil)
		if err != nil {
			return upgradeActionMsg{
				success: false,
				message: fmt.Sprintf("Failed to upgrade: %v", err),
			}
		}

		// Refresh releases
		releases, err := m.helmClient.ListReleases(true)
		if err != nil {
			return errMsg(err)
		}

		// Send both success message and refreshed releases
		return tea.Batch(
			func() tea.Msg {
				return upgradeActionMsg{
					success: true,
					message: fmt.Sprintf("Successfully upgraded to version %s", selectedVer.Version),
				}
			},
			func() tea.Msg {
				return releasesMsg(releases)
			},
		)()
	}
}
