# Helmet - Feature Reference

## Complete Feature List

### Dashboard View
**Access**: Tab key or launch application (default view)

**Features**:
- Release statistics cards
  - Total releases count
  - Deployed releases (green)
  - Failed releases (red)
  - Pending releases (yellow)
- Quick start button to launch chart catalog
- Current context and namespace display

**Keybindings**:
- `tab` - Switch to Releases view
- `s` - Quick start (open chart catalog)
- `i` - Install chart
- `q` - Quit application
- `ctrl+r` - Refresh data

---

### Releases View
**Access**: Tab key from Dashboard or Repositories

**Display Columns**:
1. **STATUS** - Visual indicator (● color + state text)
   - Green ● deployed
   - Red ● failed
   - Yellow ● pending-install/pending-upgrade
2. **NAME** - Release name (truncated at 25 chars)
3. **NAMESPACE** - K8s namespace (truncated at 20 chars)
4. **CHART** - Chart name and version (e.g., "nginx-22.2.4")
5. **APP VERSION** - Application version from chart
6. **UPDATED** - Last update time (relative format: "5m ago", "2h ago", "3d ago")

**Operations**:
- `i` - Install new chart
  - Opens installation form
  - Auto-discovers local charts in current directory
  - Supports repository charts and local paths
- `d` - Delete selected release
  - Shows "⟳ Processing..." during action
  - Shows "✓ Release deleted" on success
  - Auto-refreshes list
- `v` - View release values
  - Opens full-screen YAML viewer
  - Scrollable with ↑/↓ or j/k
- `u` - Upgrade release (planned)
- `H` - View release history (planned)
- `ctrl+r` - Refresh release list manually

**Navigation**:
- `↑/k` - Move selection up
- `↓/j` - Move selection down
- `tab` - Switch to Repositories view
- `q` - Quit application

**Visual Feedback**:
- Selected row: Kubernetes blue background
- Status messages: Green with ✓ icon
- Processing indicator: Yellow with ⟳ icon
- Messages auto-clear after 2 seconds

---

### Repositories View
**Access**: Tab key from Releases

**Display Columns**:
1. **NAME** - Repository name (e.g., "bitnami")
2. **URL** - Repository URL (truncated at 60 chars)
   - HTTP/HTTPS repositories (e.g., "https://charts.bitnami.com/bitnami")
   - Local file repositories (e.g., "file:///Users/me/charts")
   - OCI registries (e.g., "oci://mcr.microsoft.com/helm")

**Operations**:
- `a` - Add repository
  - Opens add repository form
  - Fields: Name, URL
  - Validates input before adding
  - Supports HTTP, local file://, and OCI protocols
- `r` - Remove selected repository
  - Shows "⟳ Processing..."
  - Shows "✓ Repository removed" on success
  - Auto-refreshes list
- `U` - Update all repositories (capital U)
  - Downloads latest chart information
  - Shows progress indicator
  - Updates repository indexes
- `ctrl+r` - Refresh repository list manually

**Navigation**:
- `↑/k` - Move selection up
- `↓/j` - Move selection down
- `tab` - Switch to Dashboard view
- `q` - Quit application

---

### Values Viewer
**Access**: Press `v` on any release in Releases view

**Display**:
- Full-screen YAML viewer
- Title shows release name
- Scrollable viewport with line wrapping
- Displays complete values YAML for the release
- Auto-formatted for readability

**Keybindings**:
- `↑/k` - Scroll up
- `↓/j` - Scroll down
- `esc` - Return to releases view
- `q` - Does NOT work in viewer (use esc)

**YAML Format**:
- Shows all custom values set during installation
- Displays default values from chart
- Proper indentation preserved
- Shows "No values available" if empty

---

### Install Form
**Access**: Press `i` from Dashboard or Releases view

**Form Fields**:
1. **Release Name** - Name for the Helm release (required)
2. **Chart Path** - Repository chart or local path (required)
   - Examples: `bitnami/nginx`, `./mychart`, `/path/to/chart`
   - Local charts auto-discovered and displayed
3. **Namespace** - Target namespace (default: "default")

**Features**:
- Auto-discovery of local Helm charts in current directory
- Tab/Shift+Tab to navigate between fields
- Input validation for release names (lowercase alphanumeric + hyphens)
- Visual field focus indicators
- Discovered charts shown below form

**Keybindings**:
- `tab` - Next field
- `shift+tab` - Previous field
- `enter` - Install chart (validates first)
- `esc` - Cancel installation
- Type to enter text in focused field
- `backspace` - Delete character

**Installation Process**:
1. Validates all required fields
2. Shows "⟳ Installing..." indicator
3. Executes Helm install via SDK
4. Shows success/error message
5. Returns to releases view on success
6. Auto-refreshes release list

---

### Chart Catalog
**Access**: Press `s` from Dashboard (Quick Start)

**Features**:
- Multi-select chart browser
- 10+ popular public charts pre-configured
- Automatic repository setup
- Batch installation of selected charts

**Available Charts**:
- **bitnami/nginx** - NGINX web server
- **bitnami/redis** - Redis in-memory database
- **bitnami/postgresql** - PostgreSQL database
- **bitnami/mongodb** - MongoDB NoSQL database
- **bitnami/mysql** - MySQL database
- **prometheus-community/prometheus** - Metrics collection
- **grafana/grafana** - Metrics visualization
- **ingress-nginx/ingress-nginx** - Ingress controller
- **jetstack/cert-manager** - Certificate management
- **argo/argocd** - GitOps continuous delivery

**Keybindings**:
- `↑/k` - Move selection up
- `↓/j` - Move selection down
- `space` - Toggle selection on/off
- `enter` - Install all selected charts
- `esc` - Cancel and return to dashboard

**Display**:
- Checkbox indicator for selected charts
- Repository name shown in cyan
- Chart description
- Visual selection highlighting

**Installation Behavior**:
- Automatically adds required repositories if missing
- Installs charts sequentially
- Shows progress for each chart
- Generates release names like `my-nginx`, `my-redis`
- Installs to default namespace
- Error handling for failed installations

---

### Add Repository Form
**Access**: Press `a` from Repositories view

**Form Fields**:
1. **Repository Name** - Short name identifier (required)
   - Examples: `bitnami`, `my-charts`, `azure`
2. **Repository URL** - Full URL to repository (required)
   - HTTP/HTTPS: `https://charts.bitnami.com/bitnami`
   - Local: `file:///Users/you/helm-charts`
   - OCI: `oci://mcr.microsoft.com/helm`

**Features**:
- Protocol detection (HTTP, file, OCI)
- URL validation
- Duplicate name checking
- Tab navigation between fields

**Keybindings**:
- `tab` - Next field
- `shift+tab` - Previous field
- `enter` - Add repository (validates first)
- `esc` - Cancel
- Type to enter text in focused field

**Add Process**:
1. Validates name and URL format
2. Shows "⟳ Adding repository..." indicator
3. Executes `helm repo add` via SDK
4. Updates repository index
5. Shows success/error message
6. Returns to repositories view on success
7. Auto-refreshes repository list

---

## Keyboard Reference

### Global (All Views)
| Key | Action |
|-----|--------|
| `tab` | Switch view (Dashboard → Releases → Repositories → Dashboard) |
| `q` | Quit application (except in modals) |
| `ctrl+r` | Refresh current view |

### Navigation (Lists)
| Key | Action |
|-----|--------|
| `↑` or `k` | Move up |
| `↓` or `j` | Move down |

### Release Operations
| Key | Action |
|-----|--------|
| `i` | Install new chart |
| `d` | Delete release |
| `v` | View release values (YAML) |
| `u` | Upgrade release (planned) |
| `H` | View release history (planned) |

### Repository Operations
| Key | Action |
|-----|--------|
| `a` | Add repository |
| `r` | Remove repository |
| `U` | Update all repositories |

### Dashboard Operations
| Key | Action |
|-----|--------|
| `s` | Quick start (chart catalog) |
| `i` | Install chart |

### Modal Operations
| Key | Action |
|-----|--------|
| `esc` | Close modal / Cancel action |
| `enter` | Confirm action / Submit form |
| `tab` | Next field (forms) |
| `shift+tab` | Previous field (forms) |
| `space` | Toggle selection (catalog) |

### Values Viewer / Scrolling
| Key | Action |
|-----|--------|
| `↑/k` | Scroll up |
| `↓/j` | Scroll down |
| `esc` | Exit viewer |

---

## Feature Evolution & Status

### Phase 1 MVP ✅ (Completed)
- [x] Dashboard with release statistics
- [x] Releases list view with details
- [x] Release deletion with confirmation
- [x] Tab navigation between views
- [x] Vim-style keybindings
- [x] Status bar with context/namespace
- [x] Visual feedback for operations
- [x] Kubernetes color scheme

### Phase 2 Advanced Features ✅ (Completed)
- [x] Repositories management (list, add, remove, update)
- [x] Values viewer with YAML display
- [x] Install form with local chart discovery
- [x] Chart catalog with multi-select
- [x] Quick start functionality
- [x] Support for HTTP, local, and OCI repositories
- [x] Modal overlay UI for forms and viewers
- [x] Auto-discover Helm charts in current directory

### Phase 3 Planned Features 🚧 (Roadmap)
- [ ] Release upgrade with version selection
- [ ] Release history viewer
- [ ] Rollback to previous revisions
- [ ] Kubernetes resource viewer
  - View pods, services, deployments from release
  - Status indicators for resources
- [ ] Pod logs viewer
  - Stream logs from release pods
  - Search and filter logs
- [ ] Multi-context switcher UI
  - Switch between different K8s clusters
  - Visual context selection
- [ ] Namespace switcher
  - Change active namespace interactively
  - List all available namespaces
- [ ] Search and filter functionality
  - Search releases by name
  - Filter by status
  - Search charts in catalog
- [ ] Release testing integration
  - Run Helm tests
  - Display test results

### Phase 4 Future Enhancements 💡 (Ideas)
- [ ] Custom values editor
  - Edit YAML values before install/upgrade
  - Syntax validation
  - Dry-run preview
- [ ] Chart development tools
  - Lint local charts
  - Package charts
  - Dependency management
- [ ] Bulk operations
  - Multi-select releases
  - Batch delete/upgrade
- [ ] Configuration file support
  - Save preferences (~/.helmetrc)
  - Custom color themes
  - Default namespaces
- [ ] Auto-refresh mode
  - Configurable refresh interval
  - Real-time status updates
- [ ] Export functionality
  - Export release manifests
  - Export values to files
- [ ] Helm plugin integration
  - Run installed Helm plugins
  - Plugin management

---

## Known Limitations

1. **Single Namespace View** - Currently shows releases from one namespace at a time
2. **Sync Operations** - UI blocks during Helm API calls (no async yet)
3. **No Auto-Refresh** - Must manually refresh (ctrl+r) to update data
4. **No Multi-Select Operations** - Can only act on one release at a time
5. **No Confirmation Dialogs** - Destructive operations (delete) execute immediately
6. **No Offline Mode** - Requires active connection to Kubernetes cluster
7. **Limited Error Recovery** - Some errors require application restart
8. **No Configuration File** - All settings are hardcoded
9. **No Release Upgrade** - Cannot upgrade existing releases yet
10. **No Rollback** - Cannot revert to previous release revisions

---

## Technical Implementation Notes

### Local Chart Discovery
- Recursively scans current directory for `Chart.yaml` files
- Ignores common directories: `.git`, `node_modules`, `vendor`, `target`
- Displays relative paths for easy reference
- Updates on every form open

### Repository Types Supported
- **HTTP/HTTPS**: Standard Helm chart repositories
- **file://**: Local filesystem chart repositories
- **OCI**: OCI-compliant container registries

### Status Colors
- **Green** (#50FA7B): Deployed, success states
- **Red** (#FF5555): Failed, error states
- **Yellow** (#FFB86C): Pending, warning states
- **Blue** (#326CE5): Kubernetes brand, active states
- **Purple** (#7D56F4): Selected, accent states
- **Cyan** (#8BE9FD): Info, labels

### Time Display Format
- Under 1 minute: "30s ago"
- Under 1 hour: "5m ago"
- Under 24 hours: "2h ago"
- Over 24 hours: "3d ago"

### Chart Catalog Repositories
Pre-configured repositories for quick start:
- **bitnami**: https://charts.bitnami.com/bitnami
- **prometheus-community**: https://prometheus-community.github.io/helm-charts
- **grafana**: https://grafana.github.io/helm-charts
- **ingress-nginx**: https://kubernetes.github.io/ingress-nginx
- **jetstack**: https://charts.jetstack.io
- **argo**: https://argoproj.github.io/argo-helm

---

## Performance Considerations

### Current Performance
- Fast startup (uses existing kubeconfig)
- Instant keybinding response
- List operations complete in < 1 second (typical cluster)
- Values viewer loads immediately for most releases
- Install operations depend on chart size and cluster speed

### Known Performance Bottlenecks
- Large YAML values may slow down viewer
- Many releases (100+) can cause list lag
- Repository updates can take several seconds per repo
- No caching of chart metadata

### Future Optimizations
- Async operations to prevent UI blocking
- Background polling for status updates
- Virtual scrolling for large lists
- Caching of frequently accessed data
- Incremental updates instead of full refreshes
- Lazy loading of release details
