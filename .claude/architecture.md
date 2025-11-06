# Helmet - Architecture & Design Decisions

## Architecture Overview

Helmet follows a layered architecture with clear separation between Helm/Kubernetes operations, business logic, and presentation.

```
┌─────────────────────────────────────────┐
│         User Interface Layer            │
│    (Bubble Tea Models & Views)          │
├─────────────────────────────────────────┤
│         Business Logic Layer            │
│   (State Management, Actions)           │
├─────────────────────────────────────────┤
│       Helm & Kubernetes Client          │
│    (Simplified SDK Wrapper)             │
├─────────────────────────────────────────┤
│     Helm SDK & Kubernetes API           │
└─────────────────────────────────────────┘
```

## Layer Details

### 1. Helm Client Layer (`helm/client.go`)

**Purpose**: Simplify and wrap Helm SDK v3 and Kubernetes API calls

**Key Design Decisions**:
- Single `Client` struct wrapping official Helm SDK
- Action configuration management for all Helm operations
- Namespace-aware operations
- Simplified method signatures hiding SDK complexity
- Error handling at this layer

**Example Methods**:
```go
ListReleases(namespace string) ([]*release.Release, error)
InstallChart(releaseName, chartPath, namespace string) error
UninstallRelease(releaseName, namespace string) error
GetReleaseValues(releaseName, namespace string) (map[string]interface{}, error)
AddRepository(name, url string) error
```

**Why**: Keeps UI layer clean and allows easy testing/mocking

**Helm Settings Configuration**:
- Namespace from kubeconfig or override
- Repository cache in `~/.cache/helm/repository`
- Registry config in `~/.config/helm/registry`
- Repository file in `~/.config/helm/repositories.yaml`

---

### 2. Kubernetes Layer (`kubernetes/context.go`)

**Purpose**: Manage Kubernetes cluster contexts and namespaces

**Key Components**:
```go
type ContextManager struct {
    clientConfig clientcmd.ClientConfig
    rawConfig    *api.Config
}
```

**Operations**:
- Get current context and cluster information
- List all available contexts
- Switch between contexts
- Get current namespace
- List all namespaces
- Validate kubeconfig

**Why**: Separates K8s configuration management from Helm operations

---

### 3. Chart Discovery Layer (`helm/discovery.go`)

**Purpose**: Auto-discover local Helm charts in the filesystem

**Features**:
- Recursive directory scanning for `Chart.yaml` files
- Smart filtering of ignored directories
- Relative path calculation
- Fast filesystem traversal

**Ignored Directories**:
- `.git` - Version control
- `node_modules` - Node.js dependencies
- `vendor` - Go vendor dependencies
- `target` - Build artifacts
- `.terraform` - Terraform state

**Why**: Improves UX by automatically suggesting local charts for installation

---

### 4. UI Layer (`ui/`)

**Architecture Pattern**: The Elm Architecture (TEA) via Bubble Tea

#### Main Model (`ui/model.go`)

**Responsibilities**:
- Top-level application state
- View routing (Dashboard, Releases, Repositories, Values, Forms)
- Tab navigation
- Modal management
- Window size management
- Message dispatching

**Key Components**:
```go
type Model struct {
    client       *helm.Client          // Helm operations
    currentView  View                  // Active view
    context      string                // K8s context
    namespace    string                // Active namespace
    cluster      string                // Cluster name

    // Sub-models for each view
    dashboard    *DashboardModel
    releases     *ReleasesModel
    repositories *RepositoriesModel
    values       *ValuesModel
    install      *InstallModel
    catalog      *CatalogModel
    addRepo      *AddRepoModel
}
```

**Message Flow**:
1. User input → KeyMsg or custom message
2. Model.Update() routes to active view/modal
3. View model processes message, returns commands
4. Commands execute async → send new messages
5. Model.View() renders current state

#### Sub-Models Pattern

Each view (Dashboard, Releases, etc.) implements:
- **Model struct** - View-specific state
- **NewModel() Model** - Constructor
- **Init() tea.Cmd** - Initialize and fetch data
- **Update(msg) (tea.Model, tea.Cmd)** - Handle messages
- **View() string** - Render UI
- **refresh() tea.Cmd** - Reload data

**Why**: Modularity, testability, clear ownership of state

---

## Key Design Decisions

### 1. View System Design

**Decision**: Tab-based navigation with modal overlays

**Views**:
- **Dashboard** - Release statistics, quick start
- **Releases** - List of all releases with operations
- **Repositories** - Helm repository management

**Modals** (overlays):
- **Values Viewer** - Full-screen YAML display
- **Install Form** - Chart installation interface
- **Chart Catalog** - Multi-select chart browser
- **Add Repository** - Repository configuration form

**Rationale**:
- Tabs provide clear navigation context
- Modals get full-screen for focused tasks
- Clean separation between views and actions
- Consistent UX across application

**Implementation**:
```go
// model.go View()
if m.showingValues {
    return m.values.View()
} else if m.installing {
    return m.install.View()
} else if m.showingCatalog {
    return m.catalog.View()
}

// Normal view rendering
switch m.currentView {
case DashboardView:
    content = m.dashboard.View()
case ReleasesView:
    content = m.releases.View()
case RepositoriesView:
    content = m.repositories.View()
}
return tabs + content + footer
```

---

### 2. Release View Layout

**Decision**: Tabular format with fixed-width columns

**Columns**:
- STATUS (12 chars) - Visual indicator + state
- NAME (25 chars) - Release name
- NAMESPACE (20 chars) - K8s namespace
- CHART (30 chars) - Chart name and version
- APP VERSION (15 chars) - Application version
- UPDATED (15 chars) - Relative time

**Why**:
- Predictable, clean layout
- Easy to scan and compare
- Prevents text wrapping issues
- Truncation with "..." for overflow
- Aligns with Helm CLI output format

---

### 3. Action Feedback System

**Decision**: Visual indicators for async operations

**Implementation**:
```go
type ReleasesModel struct {
    statusMsg        string  // Success/error message
    actionInProgress bool    // Loading state
}
```

**Flow**:
1. User presses `d` (delete)
2. `actionInProgress = true` → shows "⟳ Processing..."
3. Helm API call executes
4. On completion → `statusMsg = "✓ Release deleted"`
5. Auto-clear after 2 seconds
6. List refreshes with updated data

**Why**: User needs confirmation that actions are happening and completed

---

### 4. Values Viewer Design

**Challenge**: Display potentially large YAML files in terminal

**Solution**: Scrollable viewport with Bubble Tea viewport component

**Features**:
- Syntax-aware YAML display
- Vim-style scrolling (j/k, ↑/↓)
- Line wrapping for readability
- Full-screen display
- Escape to exit

**Implementation**:
```go
type ValuesModel struct {
    releaseName string
    viewport    viewport.Model
    ready       bool
}
```

**Why**: YAML values can be hundreds of lines; need efficient scrolling

---

### 5. Chart Catalog Implementation

**Decision**: Pre-configured list with multi-select

**Features**:
- 10+ popular charts from well-known repositories
- Space to toggle selection
- Visual checkbox indicators
- Automatic repository setup
- Batch installation

**Chart Selection**:
```go
type CatalogChart struct {
    Name        string  // e.g., "bitnami/nginx"
    Repo        string  // e.g., "bitnami"
    RepoURL     string  // e.g., "https://charts.bitnami.com/bitnami"
    Chart       string  // e.g., "nginx"
    Description string
}
```

**Why**: Lowers barrier to entry, provides quick start experience

---

### 6. Local Chart Discovery

**Decision**: Auto-scan current directory on form open

**Implementation**:
- Recursive search for `Chart.yaml` files
- Display relative paths in form
- Real-time discovery (no caching)
- Filtered results (ignores common directories)

**Why**: Improves developer experience when working with local charts

---

## Styling System (`ui/styles.go`)

### Color Palette
```go
k8sBlue        = "#326CE5"  // Kubernetes brand - primary
primaryColor   = "#7D56F4"  // Purple - accents, selections
successColor   = "#50FA7B"  // Green - deployed, success
warningColor   = "#FFB86C"  // Orange - pending, warnings
errorColor     = "#FF5555"  // Red - failed, errors
infoColor      = "#8BE9FD"  // Cyan - labels, info
mutedColor     = "#6272A4"  // Gray - inactive, help text
```

### Design Philosophy
- **Primary**: Kubernetes blue for brand recognition
- **Status Colors**: Semantic (green=good, red=bad, yellow=caution)
- **Contrast**: High contrast for terminal readability
- **Consistency**: Same colors across all views

### Style Patterns
- **ActiveTabStyle**: Bold, white text, K8s blue background
- **InactiveTabStyle**: Muted gray text, no background
- **CardStyle**: Rounded borders, padding, margin
- **StatusStyles**: Color-coded by state (deployed/failed/pending)
- **SelectedItemStyle**: K8s blue background for list selections

### Border Usage
- Tabs: No borders (clean, modern look)
- Cards: Rounded borders for dashboard statistics
- Modals: Thick rounded borders to indicate overlay
- Title: Bottom border for section separation

---

## Message Patterns

### Custom Message Types
```go
// Data fetch messages
type releasesMsg []*release.Release
type repositoriesMsg []*repo.Entry
type valuesMsg map[string]interface{}

// Action messages
type actionSuccessMsg struct {
    message string
}

type actionErrorMsg struct {
    error error
}

// UI control messages
type clearStatusMsg struct{}
type windowSizeMsg tea.WindowSizeMsg
```

### Async Patterns
```go
// Command that returns message
func (m *ReleasesModel) refresh() tea.Cmd {
    return func() tea.Msg {
        releases, err := m.client.ListReleases(m.namespace)
        if err != nil {
            return actionErrorMsg{err}
        }
        return releasesMsg(releases)
    }
}
```

**Why**: Bubble Tea's message-driven architecture enables async operations without blocking UI

---

## Error Handling

### Strategy
1. **API Layer**: Return Go errors with context
2. **Model Layer**: Convert to `actionErrorMsg` type
3. **View Layer**: Display with `ErrorStyle`
4. **Recovery**: Clear error on next action

### User Experience
- Errors shown in red with border
- Operations continue (no crashes)
- Error cleared on next successful action
- Actionable error messages when possible

**Example**:
```go
if err != nil {
    return m, func() tea.Msg {
        return actionErrorMsg{
            error: fmt.Errorf("failed to install chart: %w", err),
        }
    }
}
```

---

## Performance Considerations

### Current Implementation
- Synchronous operations (blocking UI)
- Full refresh on actions
- In-memory YAML storage
- No caching of release data
- Direct Helm SDK calls

### Design Trade-offs
**Chosen**: Simple, synchronous operations
**Alternative**: Background workers, incremental updates
**Rationale**:
- Simpler code, easier to reason about
- Most operations complete in < 1 second
- Reliability over performance for MVP
- Easy to optimize later

### Future Improvements
- Background polling for release status
- Incremental list updates
- Caching of repository metadata
- Virtual scrolling for large lists
- Async operations with loading spinners
- Debounced refresh operations

---

## State Management

### Application State
```go
type Model struct {
    // Identity
    context   string  // Current K8s context
    namespace string  // Active namespace
    cluster   string  // Cluster name

    // Navigation
    currentView View  // Active primary view

    // Modal State
    showingValues   bool  // Values viewer open
    installing      bool  // Install form open
    showingCatalog  bool  // Catalog open
    addingRepo      bool  // Add repo form open

    // Sub-models (view-specific state)
    dashboard    *DashboardModel
    releases     *ReleasesModel
    repositories *RepositoriesModel
    values       *ValuesModel
    install      *InstallModel
    catalog      *CatalogModel
    addRepo      *AddRepoModel
}
```

### State Flow
1. User action (key press)
2. Main model routes to active view
3. View model updates its state
4. View model may dispatch async command
5. Command completes, sends message
6. Model processes message, updates state
7. View re-renders with new state

**Why**: Unidirectional data flow, predictable state changes

---

## Testing Strategy (Future)

### Planned
- Unit tests for Helm client wrapper
- Mock Helm SDK for UI testing
- Snapshot tests for view rendering
- Integration tests with test cluster (kind)
- Table-driven tests for keybindings

### Challenges
- Bubble Tea is inherently stateful
- Terminal rendering hard to test
- Async commands need careful mocking
- Helm SDK integration requires cluster

### Approach
- Test business logic separately from UI
- Mock Helm client interface
- Test message handlers in isolation
- Use golden files for view output

---

## Security Considerations

### Kubeconfig Access
- Read-only access to kubeconfig
- Respects user's context and namespace settings
- No credential storage
- Uses Kubernetes RBAC for permissions

### Helm Operations
- All operations use user's Kubernetes credentials
- No privilege escalation
- Respects namespace boundaries
- Audit trail via Helm release history

### Input Validation
- Release names: lowercase alphanumeric + hyphens
- Repository URLs: protocol validation (HTTP/file/OCI)
- Namespace names: valid K8s identifiers
- Chart paths: filesystem and repo validation

---

## Extensibility

### Adding New Views
1. Create view model struct in `ui/`
2. Implement Init(), Update(), View() methods
3. Add view type to main model enum
4. Wire up in main model's Update() and View()
5. Add keybinding in keys.go

### Adding New Operations
1. Add method to `helm/client.go`
2. Create custom message type
3. Add command in view model
4. Handle message in Update()
5. Add keybinding

### Adding New Charts to Catalog
1. Add repository to catalog model
2. Add chart entry with name, repo, description
3. Chart will auto-install on selection

---

## Code Organization

### File Structure
```
helmet/
├── main.go              # Entry point, program setup
├── helm/
│   ├── client.go        # Helm SDK wrapper (295 lines)
│   └── discovery.go     # Local chart discovery (55 lines)
├── kubernetes/
│   └── context.go       # Context management (120 lines)
└── ui/
    ├── model.go         # Main application model (700+ lines)
    ├── dashboard.go     # Dashboard view (110 lines)
    ├── releases.go      # Releases view (210 lines)
    ├── repositories.go  # Repositories view (175 lines)
    ├── values.go        # Values viewer (115 lines)
    ├── install.go       # Install form (200 lines)
    ├── catalog.go       # Chart catalog (200 lines)
    ├── addrepo.go       # Add repository form (175 lines)
    ├── styles.go        # UI styling (170 lines)
    └── keys.go          # Keybindings (130 lines)
```

### Design Principles
- **Separation of Concerns**: UI, business logic, and API layers
- **Single Responsibility**: Each file has one clear purpose
- **Composition**: Views composed from sub-models
- **Encapsulation**: Internal state not exposed
- **Testability**: Interfaces for mocking

---

## Dependencies Architecture

### Core TUI Stack
```
bubbletea (TUI framework)
    ↓
lipgloss (styling) + bubbles (components)
    ↓
Terminal rendering
```

### Helm/K8s Stack
```
helm/client.go
    ↓
Helm SDK (helm.sh/helm/v3)
    ↓
Kubernetes client-go
    ↓
K8s API
```

### Minimal External Dependencies
- No web frameworks
- No database
- No additional configuration
- Uses existing kubeconfig

**Why**: Reduces complexity, improves reliability, faster startup

---

## Evolution & Iterations

### Issues Fixed During Development

1. **Helm SDK Compatibility**
   - Problem: `repo.NewChartRepository()` signature changed
   - Solution: Updated to use `getter.All(settings)` for providers

2. **Chart Type Confusion**
   - Problem: Wrong type returned from `loadChart()`
   - Solution: Use `*chart.Chart` from `helm.sh/helm/v3/pkg/chart`

3. **Time Type Mismatch**
   - Problem: Helm uses custom time wrapper
   - Solution: Access via `.Time` property

4. **Local Chart Loading**
   - Problem: Difficult to load charts from filesystem
   - Solution: Auto-discovery with Chart.yaml scanning

5. **Modal Overlay Logic**
   - Problem: Modal state tracking complex
   - Solution: Boolean flags for each modal type

---

## Comparison to Docker TUI Pattern

Helmet's architecture is inspired by Docker TUI tools but adapted for Helm/Kubernetes:

| Aspect | Docker TUI | Helmet |
|--------|-----------|--------|
| **API Client** | Docker Engine SDK | Helm SDK v3 |
| **Resource Types** | Containers, Images, Networks | Releases, Charts, Repositories |
| **Context** | Docker daemon | K8s cluster + namespace |
| **Operations** | start/stop/logs | install/delete/upgrade/values |
| **State Source** | Docker daemon API | Helm release storage (K8s secrets) |
| **Configuration** | DOCKER_HOST | kubeconfig |

---

Built collaboratively with Claude Code on 2025-11-06
