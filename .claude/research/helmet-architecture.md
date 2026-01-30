# Helmet (helm-tui) Architecture Reference

## Overview

- Go TUI application (module path: `github.com/pidanou/helm-tui`), requires Go 1.22.3
- Built on the BubbleTea Model-View-Update (MVU) pattern from the Charm ecosystem
- 4 tabs: Releases, Repositories, Hub, Plugins
- Shells out to the `helm` CLI binary for all Helm operations (list, install, upgrade, delete, rollback, repo management, plugin management)
- Uses the ArtifactHub HTTP API (`https://artifacthub.io/api/v1/`) for Hub search and default value retrieval -- this is the only module that makes HTTP requests
- Flat package layout (no `internal/` directory): `releases/`, `repositories/`, `hub/`, `plugins/`, `components/`, `styles/`, `helpers/`, `types/`

## Entry Point & Boot Sequence

- **main.go**: Opens `debug.log` via `tea.LogToFile()`, stores the file handle in `helpers.LogFile`. Builds the tab labels slice from `ui.go`'s `tabLabels` (a `[]string{"Releases", "Repositories", "Hub", "Plugins"}`). Creates a BubbleTea program with `tea.WithAltScreen()` and runs it. On exit, defers closing the log file and truncating `debug.log` back to zero bytes.
- **ui.go**: `newModel(tabs)` creates a `mainModel` with 4 tab sub-models initialized via their respective `InitModel()` functions:
  - `releases.InitModel()` -- returns `(Model, tea.Cmd)` (cmd is nil)
  - `repositories.InitModel()` -- returns `(tea.Model, tea.Cmd)` (cmd is nil)
  - `hub.InitModel()` -- returns `tea.Model`
  - `plugins.InitModel()` -- returns `PluginsModel`
- `Init()` batches `createWorkingDir`, `textinput.Blink`, and all 4 sub-model `Init()` commands together via `tea.Batch()`
- `createWorkingDir` resolves `os.UserHomeDir()`, creates `~/.helm-tui/` via `os.MkdirAll()`, and stores the path in the mutable global `helpers.UserDir`. Returns `types.InitAppMsg{Err}` -- if Err is non-nil, the app quits immediately.

## Root Model (ui.go)

```go
type tabIndex uint

const (
    releasesTab     tabIndex = iota // 0
    repositoriesTab                 // 1
    hubTab                          // 2
    pluginsTab                      // 3
)

type mainModel struct {
    state      tabIndex    // active tab (0-3)
    index      int         // UNUSED field (declared, never read or written)
    width      int         // terminal width from WindowSizeMsg
    height     int         // terminal height from WindowSizeMsg
    tabs       []string    // tab label strings
    tabContent []tea.Model // 4 sub-models (one per tab)
    loaded     bool        // flipped true on InitAppMsg (guards View rendering)
}
```

### Tab Switching

- `]` moves forward (wraps from pluginsTab back to releasesTab)
- `[` moves backward (wraps from releasesTab to pluginsTab)
- `ctrl+c` quits the application

### Message Routing

- `tea.KeyMsg`: Routed to the **active tab only** (via a switch on `m.state`). The `[` and `]` keys are handled first by the root model to change tabs before the key is forwarded.
- `types.EditorFinishedMsg`: Routed to the active tab only (releases or repositories) and returns early.
- `types.InitAppMsg`: Handled by root model to set `m.loaded = true`; on error, triggers `tea.Quit`.
- `tea.WindowSizeMsg`: Forwarded to **all 4 tabs** with adjusted height (subtracting the menu bar height).
- **All other messages** (e.g., `ListReleasesMsg`, `HistoryMsg`, `PluginsListMsg`, etc.): Forwarded to **all 4 tabs** via the fallthrough at the bottom of `Update()`.

### WindowSizeMsg Bug

In the `tea.WindowSizeMsg` handler, each tab's Update is called sequentially, but the `cmd` variable is reassigned each time. Only the final assignment (plugins tab) has its `cmd` appended to the `cmds` slice. The commands returned by the first 3 tabs (releases, repositories, hub) are silently overwritten and lost:

```go
// BUG: cmd is overwritten 3 times before cmds is checked
m.tabContent[releasesTab], cmd = m.tabContent[releasesTab].Update(...)
m.tabContent[repositoriesTab], cmd = m.tabContent[repositoriesTab].Update(...)
m.tabContent[hubTab], cmd = m.tabContent[hubTab].Update(...)
m.tabContent[pluginsTab], cmd = m.tabContent[pluginsTab].Update(...)
return m, tea.Batch(cmds...) // cmds is empty; only the last cmd is implicitly lost too
```

Note: `cmds` is never appended to in this branch, so even the plugins cmd is lost. The `return m, tea.Batch(cmds...)` returns a batch of zero commands.

### View

- Shows `"loading..."` until `InitAppMsg` sets `m.loaded = true`
- Once loaded: renders `renderMenu()` (horizontal tab bar with active tab highlighted in purple) followed by the active tab's `View()`
- `renderMenu()` joins tab labels horizontally, styling the active tab with `styles.ActiveStyle.Background(styles.HighlightColor)` and inactive tabs with `styles.InactiveStyle`

## Module Pattern (Each Tab)

Every module follows a consistent multi-file pattern:

| File | Purpose |
|------|---------|
| `overview.go` / `hub.go` | Model struct definition, `InitModel()`, `Init()`, `Update()` |
| `*_commands.go` | `tea.Cmd` / `tea.Msg` functions that shell out to helm CLI or make HTTP calls |
| `*_keymap.go` | Key binding definitions (`key.Binding` structs) for help display |
| `*_view.go` | `View()` render function and sub-view renderers |

Some modules have additional sub-models with their own 4-file set (e.g., `releases/install.go`, `releases/upgrade.go`, `repositories/add.go`, `repositories/install.go`).

## Releases Module

### File Inventory

- `releases/overview.go` -- Main Model, Init, Update
- `releases/overview_commands.go` -- Helm CLI commands (list, history, delete, rollback, getNotes, getMetadata, getHooks, getValues, getManifest)
- `releases/overview_keymap.go` -- Key bindings (releasesKeys, historyKeys, readOnlyKeys)
- `releases/overview_view.go` -- View rendering (release table, detail views with tab bar)
- `releases/install.go` -- InstallModel, install wizard flow
- `releases/install_commands.go` -- installPackage, openEditorDefaultValues, searchLocalPackage, searchLocalPackageVersion, cleanValueFile
- `releases/install_keymap.go` -- installKeys (esc only)
- `releases/install_view.go` -- Install wizard View
- `releases/upgrade.go` -- UpgradeModel, upgrade wizard flow
- `releases/upgrade_commands.go` -- upgrade, openEditorWithValues, searchLocalPackage, searchLocalPackageVersion, cleanValueFile
- `releases/upgrade_keymap.go` -- upgradeKeys (esc only)
- `releases/upgrade_view.go` -- Upgrade wizard View

### Model State

```go
type selectedView int

const (
    releasesView  selectedView = iota // 0 - main release list
    historyView                       // 1 - revision history table
    notesView                         // 2 - helm get notes
    metadataView                      // 3 - helm get metadata
    hooksView                         // 4 - helm get hooks
    valuesView                        // 5 - helm get values
    manifestView                      // 6 - helm get manifest
)

type Model struct {
    selectedView selectedView      // current active view (0-6)
    keys         []keyMap          // 7-element slice, one keyMap per view
    help         help.Model        // help bar renderer
    releaseTable table.Model       // main releases table
    historyTable table.Model       // revision history table
    notesVP      viewport.Model    // notes viewport
    metadataVP   viewport.Model    // metadata viewport
    hooksVP      viewport.Model    // hooks viewport
    valuesVP     viewport.Model    // values viewport
    manifestVP   viewport.Model    // manifest viewport
    installModel InstallModel      // nested install wizard sub-model
    installing   bool              // true when install wizard is active
    upgradeModel UpgradeModel      // nested upgrade wizard sub-model
    upgrading    bool              // true when upgrade wizard is active
    deleting     bool              // true when delete confirmation is shown
    width        int               // terminal width
    height       int               // terminal height
}
```

### Release Table Columns

| Column | Sizing |
|--------|--------|
| Name | FlexFactor: 1 |
| Namespace | FlexFactor: 1 |
| Revision | Fixed Width: 10 |
| Updated | Fixed Width: 36 |
| Status | FlexFactor: 1 |
| Chart | FlexFactor: 1 |
| App version | FlexFactor: 1 |

### History Table Columns

| Column | Sizing |
|--------|--------|
| Revision | FlexFactor: 1 |
| Updated | Fixed Width: 36 |
| Status | FlexFactor: 1 |
| Chart | FlexFactor: 1 |
| App version | FlexFactor: 1 |
| Description | FlexFactor: 1 |

### Key Bindings

**releasesView (releasesKeys)**:
- `i` -- Install new release (opens install wizard)
- `D` -- Delete release (shows y/n confirmation)
- `u` -- Upgrade release (opens upgrade wizard, pre-fills release name and namespace)
- `r` -- Refresh (re-fetches release list)
- `enter` / `space` -- Show details (transitions to historyView, shrinks release table to single selected row)
- `esc` -- No-op in releasesView

**historyView (historyKeys)**:
- `i` -- Install new release
- `D` -- Delete release
- `u` -- Upgrade release
- `R` -- Rollback to selected revision
- `h` / `l` / `left` / `right` -- Cycle through detail sub-tabs (History, Notes, Metadata, Hooks, Values, Manifest) with wrapping
- `esc` -- Back to releasesView (restores full release table from cache)

**readOnlyViews (readOnlyKeys)** -- used for notesView, metadataView, hooksView, valuesView, manifestView:
- Same as historyKeys except no Rollback binding
- `i` -- Install, `D` -- Delete, `u` -- Upgrade
- `h` / `l` / `left` / `right` -- Navigate sub-tabs
- `esc` -- Back to releasesView

### Detail Tab Bar

When a release is selected (enter/space), the view changes to a split layout:
1. Top: Collapsed release table (3-row height, single selected row)
2. Middle: Tab bar with labels: History, Notes, Metadata, Hooks, Values, Manifest (styled with rounded borders, active tab highlighted)
3. Bottom: Content area (table or viewport depending on selected sub-tab)

Navigation wraps: from manifestView, pressing `l`/`right` goes to historyView; from historyView, pressing `h`/`left` goes to manifestView.

### Release Table Caching

A package-level `var releaseTableCache table.Model` is used to save the full release table when entering detail view. When the user presses `enter`/`space`, the current table is cached, then the release table is collapsed to show only the selected row. On `esc`, the cached table is restored.

### Helm Commands (overview_commands.go)

| Function | Command | Returns |
|----------|---------|---------|
| `list()` | `helm ls --all-namespaces --output json` | `types.ListReleasesMsg` |
| `history()` | `helm history <name> --namespace <ns> --output json` | `types.HistoryMsg` |
| `delete()` | `helm uninstall <name> --namespace <ns>` | `types.DeleteMsg` |
| `rollback()` | `helm rollback <name> <revision> --namespace <ns>` | `types.RollbackMsg` |
| `getNotes()` | `helm get notes <name> --namespace <ns>` | `types.NotesMsg` |
| `getMetadata()` | `helm get metadata <name> --namespace <ns>` | `types.MetadataMsg` |
| `getHooks()` | `helm get hooks <name> --namespace <ns>` | `types.HooksMsg` |
| `getValues()` | `helm get values <name> --namespace <ns>` | `types.ValuesMsg` |
| `getManifest()` | `helm get manifest <name> --namespace <ns>` | `types.ManifestMsg` |

Note: `getValues()` strips the first line of output (which is typically `USER-SUPPLIED VALUES:` header) before returning content.

### Data Flow on ListReleasesMsg

When `ListReleasesMsg` is received, the Update handler:
1. Sets rows on the release table (if in releasesView)
2. Caches the table with current rows and columns into `releaseTableCache`
3. Fires 6 additional commands in parallel: `m.history`, `m.getNotes`, `m.getMetadata`, `m.getHooks`, `m.getValues`, `m.getManifest`

This means detail data is pre-fetched for the first/selected release immediately upon loading.

### Install Wizard (6 steps)

```go
const (
    installChartReleaseNameStep int = iota // 0
    installChartNameStep                   // 1
    installChartVersionStep                // 2
    installChartNamespaceStep              // 3
    installChartValuesStep                 // 4
    installChartConfirmStep                // 5
)
```

**Step prompts:**
1. "Enter release name" -- free text
2. "Enter chart" -- text input with autocomplete suggestions (debounced 500ms via `types.DebounceEndMsg` with tag-based deduplication)
3. "Enter chart version (empty for latest)" -- text input with autocomplete suggestions (debounced)
4. "Enter namespace (empty for default)" -- free text, defaults to "default" if empty
5. "Edit default values ? y/n" -- if `y`, opens `$EDITOR` (or `vim` as fallback) with chart default values written to `~/.helm-tui/<namespace>/<release>/values.yaml`; if `n`, skips to confirm
6. "Enter to install" -- pressing enter triggers `installPackage()`

**Install Commands (install_commands.go):**
- `installPackage(mode)`: Runs `helm install <name> <chart> --version <ver> --namespace <ns> --create-namespace` (with `--values <file>` if mode is "y")
- `openEditorDefaultValues()`: Runs `helm show values <chart> --version <ver>`, writes output to file, then opens via `helpers.WriteAndOpenFile()` using `tea.ExecProcess()`
- `searchLocalPackage()`: Runs `helm search repo <query> --output json` -- returns package name suggestions synchronously
- `searchLocalPackageVersion()`: Runs `helm search repo --regexp \v<chart>\v --versions --output json` -- returns version suggestions synchronously
- `cleanValueFile(folder)`: Removes the temp values folder via `os.RemoveAll()`

**Debounce mechanism:** Each keystroke increments a `tag` counter and schedules a `tea.Tick` for 500ms. When the tick fires, it sends `DebounceEndMsg{Tag: tag}`. The handler only processes the message if `msg.Tag == m.tag` (meaning no new keystrokes occurred since the tick was scheduled).

### Upgrade Wizard (4 steps)

```go
const (
    upgradeReleaseChartStep   int = iota // 0
    upgradeReleaseVersionStep            // 1
    upgradeReleaseValuesStep             // 2
    upgradeReleaseConfirmStep            // 3
)
```

**Step prompts:**
1. "Enter a chart name or chart directory (absolute path)" -- autocomplete with debounce
2. "Version (empty for latest)" -- autocomplete with debounce
3. "Edit values yes/no/use default ? y/n/d" -- `y` opens editor with current release values (`helm get values`), `d` opens editor with chart default values (`helm show values`), `n` skips
4. "Confirm ? enter/esc"

**Pre-filled data:** `ReleaseName` and `Namespace` are set from the selected release table row before the upgrade wizard is activated.

**Upgrade Commands (upgrade_commands.go):**
- `upgrade()`: Runs `helm upgrade <name> <chart> --namespace <ns>` (with `--values <file>` if values were edited)
- `openEditorWithValues(defaultValues bool)`: If `defaultValues` is true, runs `helm show values <chart> --version <ver>`; otherwise runs `helm get values <name> --namespace <ns>` to get current values. Writes to file and opens editor.
- `searchLocalPackage()` and `searchLocalPackageVersion()`: Same logic as install commands (duplicated code)

## Repositories Module

### File Inventory

- `repositories/overview.go` -- Main Model, Init, Update
- `repositories/overview_commands.go` -- Helm CLI commands (list, update, remove, searchPackages, searchPackageVersions, getDefaultValue)
- `repositories/overview_keymap.go` -- Key bindings (repoListKeys, chartsListKeys, versionsKeys, defaultValuesKeyHelp)
- `repositories/overview_view.go` -- View rendering (3-panel layout)
- `repositories/add.go` -- AddModel, add repo wizard
- `repositories/add_commands.go` -- addRepo command, input helpers
- `repositories/add_keymap.go` -- addKeys (esc only)
- `repositories/add_view.go` -- Add wizard View
- `repositories/install.go` -- InstallModel, install from repo wizard
- `repositories/install_commands.go` -- installPackage, openEditorDefaultValues, cleanValueFile
- `repositories/install_keymap.go` -- installKeys (esc only)
- `repositories/install_view.go` -- Repo install wizard View

### Model State

```go
type selectedView int

const (
    listView     selectedView = iota // 0 - repository list (left panel)
    packagesView                     // 1 - packages in selected repo (middle panel)
    versionsView                     // 2 - versions of selected package (right panel)
)

type Model struct {
    selectedView     selectedView      // active panel (0-2)
    keys             []keyMap          // 3-element slice, one per panel
    tables           []table.Model     // 3 tables: repos, packages, versions
    installModel     InstallModel      // nested install wizard
    addModel         AddModel          // nested add repo wizard
    help             help.Model        // help bar
    installing       bool              // install wizard active
    adding           bool              // add repo wizard active
    defaultValueVP   viewport.Model    // viewport for chart default values
    showDefaultValue bool              // show default value overlay
    width            int
    height           int
}
```

### Layout

The repositories view uses a 3-panel side-by-side layout:

- **Left panel** (1/4 width): Repository list table
  - Columns: Name (flex 1), URL (flex 3)
- **Middle panel** (1/4 width): Packages in selected repository
  - Columns: Name (flex 1)
- **Right panel** (2/4 width): Versions of selected package
  - Columns: Chart Version (fixed 13), App Version (fixed 13), Description (flex 1)

The active panel is highlighted with `styles.HighlightColor` on the top border and `styles.ActiveStyle` on the side/bottom borders.

### Key Bindings

**All panels share:**
- `D` -- Delete selected repo
- `u` -- Update selected repo
- `i` -- Install selected version (requires package and version selected)
- `a` -- Add new repo (opens add wizard)
- `v` -- Show default values for selected package/version
- `r` -- Refresh (re-fetches repo list)
- `h` / `l` / `left` / `right` -- Switch between panels (no wrapping: left stops at listView, right stops at versionsView)
- `j` / `k` / `up` / `down` -- Navigate within current panel; triggers cascading data refresh (selecting a repo refreshes packages, selecting a package refreshes versions)
- `esc` -- Close overlays (install wizard, add wizard, default value view), return to listView

### Cascading Data Refresh

When navigating up/down within a panel:
- In `listView`: Triggers `searchPackages` to update packages for the newly selected repo
- In `packagesView`: Triggers `searchPackageVersions` to update versions for the newly selected package

This creates a drill-down experience: Repository -> Packages -> Versions.

### Helm Commands (overview_commands.go)

| Function | Command(s) | Returns |
|----------|-----------|---------|
| `list()` | `helm repo update` then `helm repo ls --output json` | `types.ListRepoMsg` |
| `update()` | `helm repo update <name>` | `types.UpdateRepoMsg` |
| `remove()` | `helm repo remove <name>` | `types.RemoveMsg` |
| `searchPackages()` | `helm search repo <repoName>/ --output json` | `types.PackagesMsg` |
| `searchPackageVersions()` | `helm search repo <pkgName> --versions --output json` | `types.PackageVersionsMsg` |
| `getDefaultValue()` | `helm show values <pkgName> --version <ver>` | `types.DefaultValueMsg` |

Note: The `list()` function always runs `helm repo update` first before listing, which means every refresh triggers a network call to update all repo indices.

### Add Repo Wizard (2 steps)

```go
const (
    repoNameStep int = iota // 0
    urlStep                 // 1
)
```

1. "Enter repo name" -- free text
2. "Enter repo URL" -- free text, enter triggers `helm repo add <name> <url>`

### Install from Repo Wizard (4 steps)

```go
const (
    nameStep      installStep = iota // 0
    namespaceStep                    // 1
    valuesStep                       // 2
    confirmStep                      // 3
)
```

1. "Enter release name" -- free text
2. "Enter namespace (empty for default)" -- free text
3. "Edit default values ? y/n" -- opens editor with `helm show values` output if `y`
4. "Enter to install" -- triggers `helm install`

Pre-filled: `Chart` and `Version` are set from the selected package/version table rows before the wizard is activated.

## Hub Module

### File Inventory

- `hub/hub.go` -- HubModel, InitModel, Init, Update
- `hub/hub_commands.go` -- HTTP API commands (searchHub, searchDefaultValue) and helm CLI (addRepo)
- `hub/hub_keymap.go` -- Key bindings (defaultKeysHelp, tableKeysHelp, searchKeyHelp, addRepoKeyHelp, defaultValuesKeyHelp)
- `hub/hub_view.go` -- View rendering

### Model State

```go
type HubModel struct {
    searchBar      textinput.Model   // search input at the top
    resultTable    table.Model       // search results table
    defaultValueVP viewport.Model    // viewport for default values display
    repoAddInput   textinput.Model   // input for local repo name when adding
    help           help.Model        // help bar
    width          int
    height         int
    view           int               // 0=searchView, 1=defaultValueView
}

const (
    searchView       int = iota // 0
    defaultValueView            // 1
)
```

### Result Table Columns

| Column | Sizing | Notes |
|--------|--------|-------|
| id | Fixed Width: 0 | Hidden -- stores ArtifactHub package_id |
| version | Fixed Width: 0 | Hidden -- stores package version |
| Package | FlexFactor: 1 | Visible |
| Repository | FlexFactor: 1 | Visible |
| URL | FlexFactor: 3 | Visible |
| Description | FlexFactor: 3 | Visible |

The `id` and `version` columns have width 0, making them invisible in the UI but accessible via `SelectedRow()` for API calls.

### Key Bindings

**Default state (no focus):**
- `/` -- Focus the search bar
- `enter` -- Focus the result table

**Search bar focused:**
- `enter` -- Execute search and focus table

**Result table focused:**
- `v` -- Show default values for selected package
- `/` -- Focus search bar
- `a` -- Show repo-add input (prompts for local repository name)

**Repo add input focused:**
- `enter` -- Execute `helm repo add`

**Default value view:**
- `/` -- Focus search bar
- `esc` -- Return to search view

### Commands (hub_commands.go) -- HTTP API

This is the **only module** that makes HTTP requests (all other modules exclusively shell out to `helm` CLI).

| Function | Method | URL | Returns |
|----------|--------|-----|---------|
| `searchHub()` | GET | `https://artifacthub.io/api/v1/packages/search?offset=0&limit=20&facets=false&ts_query_web=<query>&kind=0&deprecated=false&sort=relevance` | `types.HubSearchResultMsg` |
| `searchDefaultValue()` | GET | `https://artifacthub.io/api/v1/packages/<id>/<version>/values` | `types.HubSearchDefaultValueMsg` |
| `addRepo()` | N/A (helm CLI) | `helm repo add <localName> <url>` | `types.AddRepoMsg` |

**searchHub()** parses a `Response` struct containing `[]Package` where each Package has: `ID`, `NormalizedName`, `Description`, `Version`, and nested `Repository{Name, URL}`. Results are limited to 20, sorted by relevance, Helm charts only (`kind=0`), excluding deprecated.

**searchDefaultValue()** requests YAML content (`Accept: application/yaml`) from the ArtifactHub values endpoint using the package ID and version from the hidden table columns.

**addRepo()** uses the URL from column index 4 of the selected result row and the user-provided local name from `repoAddInput`.

### View Layout

- Top: Search bar (bordered, highlighted when focused)
- Middle: Results table with " Results " title border (highlighted when focused)
- Bottom (conditional): Repo add input (shown only when `repoAddInput` is focused)
- Help bar at the bottom

When `view == defaultValueView`, the entire view is replaced with a full-screen viewport showing default values with a " Default Values " title border.

## Plugins Module

### File Inventory

- `plugins/overview.go` -- PluginsModel, InitModel, Init, Update
- `plugins/overview_commands.go` -- Helm CLI commands (list, install, update, uninstall)
- `plugins/overview_keymap.go` -- Key bindings (overviewKeys)
- `plugins/overview_view.go` -- View rendering

### Model State

```go
type PluginsModel struct {
    pluginsTable       table.Model       // plugins list table
    installPluginInput textinput.Model    // text input for plugin path/URL
    help               help.Model        // help bar
    keys               keyMap            // single keyMap (simplest module)
    width              int
    height             int
}
```

This is the simplest module -- a single view with a single table and an optional install input overlay.

### Plugins Table Columns

| Column | Sizing |
|--------|--------|
| Name | FlexFactor: 1 |
| Version | FlexFactor: 1 |
| description | FlexFactor: 3 |

Note: The "description" column title is lowercase (inconsistent with other modules that use title case).

### Key Bindings (overviewKeys)

- `i` -- Install (shows the install input field, `installPluginInput.Focus()`)
- `u` -- Update selected plugin (only works when install input is not focused)
- `U` -- Uninstall selected plugin (only works when install input is not focused)
- `r` -- Refresh (re-fetches plugin list)
- `esc` -- Cancel (blurs install input)
- `enter` -- When install input is focused, triggers install

### Helm Commands (overview_commands.go)

| Function | Command | Returns |
|----------|---------|---------|
| `list()` | `helm plugin ls` | `types.PluginsListMsg` |
| `install()` | `helm plugin install <path/url>` | `types.PluginInstallMsg` |
| `update()` | `helm plugin update <name>` | `types.PluginUpdateMsg` |
| `uninstall()` | `helm plugin uninstall <name>` | `types.PluginUninstallMsg` |

**list() parsing**: Unlike other modules that use `--output json`, `helm plugin ls` outputs tab-delimited text. The parser skips the header line (`lines[1:]`), splits each line by whitespace via `strings.Fields()`, takes `fields[0]` as name, `fields[1]` as version, and joins `fields[2:]` as description.

**Note**: The `list()` function erroneously returns `types.ListReleasesMsg{Err: err}` on error (a releases message type instead of `types.PluginsListMsg`). This does not cause a crash because BubbleTea routes messages by type, but the error message will be received by the releases module instead.

### View Layout

- Main: Plugins table rendered via `components.RenderTable()` (this is the only module that uses the shared `RenderTable` function)
- Conditional: Install input field below the table (shown when `installPluginInput` is focused)
- Bottom: Help bar

## Shared Infrastructure

### components/table.go

**ColumnDefinition** struct:
```go
type ColumnDefinition struct {
    Title      string
    Width      int    // fixed pixel width (0 means use FlexFactor)
    FlexFactor int    // proportional width allocation (0 means use Width)
}
```

**SetTable(t, cols, targetWidth)**: Flex layout algorithm for column widths.
1. Subtracts 2 from targetWidth for borders
2. Calculates total flex factor and remaining width after fixed columns
3. Allocates flex columns proportionally: `remainingWidth * flexFactor / totalFlex - 2` (minus cell padding)
4. Adjusts the last column to absorb any rounding remainder from integer division

**GenerateTable()**: Factory function that creates a `table.Model` with custom styling:
- Header: Rounded border bottom, foreground color "240", bold
- Selected row: Foreground "229" (cream), background "57" (purple), not bold
- Unbinds: `HalfPageUp`, `PageDown`, `HalfPageDown` (unbound **twice** -- duplicate call bug), `GotoBottom`, `GotoTop`

**RenderTable(t, height, width)**: Renders a table with a titled top border.
- **BUG**: Hardcodes the title as `" Releases "` for ALL tabs that use this function. Currently only the plugins module uses `RenderTable()`, so the plugins table incorrectly shows " Releases " as its title.

### styles/

**styles.go:**
```go
var (
    InactiveTabBorder = tabBorderWithBottom("...", "...", "...")  // Modified rounded border for tabs
    ActiveTabBorder   = tabBorderWithBottom("...", " ", "...")    // Active tab has space bottom
    InactiveTabStyle  = lipgloss.NewStyle().Border(InactiveTabBorder, true).Padding(0, 1)
    ActiveTabStyle    = InactiveTabStyle.Border(ActiveTabBorder, true)
    WindowSize        tea.WindowSizeMsg  // Global mutable (inconsistently used)
    Border            = lipgloss.Border(lipgloss.RoundedBorder())
    HighlightColor    = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
    InactiveStyle     = lipgloss.NewStyle()
    ActiveStyle       = InactiveStyle.BorderForeground(HighlightColor)
)
```

- `Border` is the standard rounded border used throughout the app
- `HighlightColor` is adaptive: `#874BFD` on light backgrounds, `#7D56F4` on dark backgrounds
- `ActiveStyle` adds purple border foreground to `InactiveStyle`
- `WindowSize` is a global mutable `tea.WindowSizeMsg` that is declared but inconsistently used across the codebase (modules track their own width/height instead)
- `ActiveTabStyle` / `InactiveTabStyle` are used for the detail sub-tab bar in the releases module

**helpers.go:**
- `GenerateTopBorderWithTitle(title, width, border, style)`: Creates a top border string with a centered title. Calculates left/right padding of border runes to center the title within the given width. Used by all modules to create titled panels.

### helpers/

**keymaps.go:**
```go
var CommonKeys = keyMap{
    MenuNext: key.NewBinding(key.WithKeys("[", "]"), key.WithHelp("[/]", "Change panel")),
    Quit:     key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "Quit")),
}

var SuggestionInputKeyMap = SuggestionKeyMap{
    AcceptSuggestion: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "Accept suggestion")),
    NextSuggestion:   key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("down/ctrl+n", "Next suggestion")),
    PrevSuggestion:   key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("up/ctrl+p", "Previous suggestion")),
}
```

`CommonKeys` is shown in the help bar of every module. `SuggestionInputKeyMap` is shown alongside install/upgrade wizards when autocomplete inputs are focused.

**constants.go:**
```go
var UserDir string
```
Mutable global set to `~/.helm-tui/` during initialization. Used by install/upgrade wizards to store temporary `values.yaml` files at `~/.helm-tui/<namespace>/<release>/values.yaml`.

**logging.go:**
```go
var LogFile *os.File

func Println(args ...any) {
    args = append(args, "\n")
    fmt.Fprint(LogFile, args...)
}
```
Simple file-based debug logger. `LogFile` is set in `main.go` and the file is truncated on exit.

**editor.go:**
```go
func WriteAndOpenFile(content []byte, file string) tea.Cmd
```
Writes content to a file and opens it in `$EDITOR` (falls back to `vim`). Uses `tea.ExecProcess()` to hand control to the external editor process, returning `types.EditorFinishedMsg` when the editor exits.

### types/

**helm.go** -- Data structures for helm CLI JSON output:

```go
type Pkg struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    AppVersion  string `json:"app_version"`
    Description string `json:"description"`
}

type Release struct {
    Name       string `json:"name"`
    Namespace  string `json:"namespace"`
    Revision   string `json:"revision"`    // NOTE: string type
    Updated    string `json:"updated"`
    Status     string `json:"status"`
    Chart      string `json:"chart"`
    AppVersion string `json:"app_version"`
}

type History struct {
    Revision    int    `json:"revision"`   // NOTE: int type (inconsistent with Release.Revision)
    Updated     string `json:"updated"`
    Status      string `json:"status"`
    Chart       string `json:"chart"`
    AppVersion  string `json:"app_version"`
    Description string `json:"description"`
}

type Repository struct {
    Name string `json:"name"`
    URL  string `json:"url"`
}

type Plugin struct {
    Name        string `json:"name"`
    Version     string `json:"version"`
    Description string `json:"description"`
}
```

**messages.go** -- 22 BubbleTea message types:

| Message Type | Fields | Used By |
|-------------|--------|---------|
| `InitAppMsg` | Err | Root model (app init) |
| `DeleteMsg` | Err | Releases (uninstall) |
| `ListReleasesMsg` | Content ([]table.Row), Err | Releases (list) |
| `HistoryMsg` | Content ([]table.Row), Err | Releases (history) |
| `RollbackMsg` | Err | Releases (rollback) |
| `UpgradeMsg` | Err | Releases (upgrade) |
| `NotesMsg` | Content (string), Err | Releases (get notes) |
| `MetadataMsg` | Content (string), Err | Releases (get metadata) |
| `HooksMsg` | Content (string), Err | Releases (get hooks) |
| `ValuesMsg` | Content (string), Err | Releases (get values) |
| `ManifestMsg` | Content (string), Err | Releases (get manifest) |
| `RemoveMsg` | Err | Repositories (remove) |
| `ListRepoMsg` | Content ([]table.Row), Err | Repositories (list) |
| `PackagesMsg` | Content ([]table.Row), Err | Repositories (search packages) |
| `PackageVersionsMsg` | Content ([]table.Row), Err | Repositories (search versions) |
| `InstallMsg` | Err | Releases + Repositories (install) |
| `EditorFinishedMsg` | Err | Releases + Repositories (editor) |
| `AddRepoMsg` | Err | Repositories + Hub (add repo) |
| `UpdateRepoMsg` | Err | Repositories (update) |
| `DebounceEndMsg` | Tag (int) | Releases install/upgrade (debounce) |
| `HubSearchResultMsg` | Content ([]table.Row), Err | Hub (search) |
| `HubSearchDefaultValueMsg` | Content (string), Err | Hub (default values) |
| `DefaultValueMsg` | Content (string), Err | Repositories (default values) |
| `PluginsListMsg` | Content ([]table.Row), Err | Plugins (list) |
| `PluginInstallMsg` | Err | Plugins (install) |
| `PluginUpdateMsg` | Err | Plugins (update) |
| `PluginUninstallMsg` | Err | Plugins (uninstall) |

Note: Some message types are shared across modules (e.g., `InstallMsg` is used by both releases and repositories install wizards, `AddRepoMsg` is used by both repositories and hub). Because the root model forwards non-key messages to all tabs, a message intended for one module may be received by another. This works correctly because each module's Update function only handles the message types it cares about.

## Known Bugs

1. **ui.go WindowSizeMsg**: The `cmd` variable is overwritten sequentially for all 4 tabs, and the `cmds` slice is never appended to. All resize commands from all tabs are lost. The `return m, tea.Batch(cmds...)` returns an empty batch.

2. **components/table.go RenderTable**: Hardcodes `" Releases "` as the title for ALL tables rendered through this function. Currently only the plugins module calls `RenderTable()`, so the plugins tab incorrectly shows "Releases" as its table title.

3. **components/table.go GenerateTable**: `k.HalfPageDown.Unbind()` is called twice in succession (duplicate line). The second call is a no-op but indicates a copy-paste error -- likely one should be `k.PageUp.Unbind()` or similar.

4. **ui.go mainModel.index field**: The `index int` field in the `mainModel` struct is declared but never read, written, or used anywhere in the codebase.

5. **types/helm.go Revision type inconsistency**: `History.Revision` is `int` but `Release.Revision` is `string`. Both represent Helm revision numbers. This works because the history command output has integer revisions in JSON while the list command has string revisions, but it creates an inconsistency in the type system.

6. **plugins/overview_commands.go list() error type**: The `list()` function returns `types.ListReleasesMsg{Err: err}` on error instead of `types.PluginsListMsg{Err: err}`. This means plugin list errors are sent as release list messages.

## Dependencies

### Direct Dependencies (go.mod)

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/charmbracelet/bubbletea` | v1.2.4 | TUI framework (MVU pattern) |
| `github.com/charmbracelet/bubbles` | v0.20.0 | TUI components (table, viewport, textinput, help) |
| `github.com/charmbracelet/lipgloss` | v1.0.0 | Terminal styling and layout |
| `github.com/stretchr/testify` | v1.10.0 | Test assertions |

### Runtime Requirements

- **Helm 3** must be installed and available in `PATH` (all operations shell out to `helm`)
- `$EDITOR` environment variable (optional, falls back to `vim`) for values editing
- Network access for ArtifactHub API (Hub tab only)
- Network access for Helm repo operations (repo update/add)

### Build & Release

**Makefile targets:**

| Target | Description |
|--------|-------------|
| `build` | Clean, format, then `go build -o ./bin/helm-tui .` |
| `run` | Build then execute `./bin/helm-tui` |
| `test` | `go test ./...` |
| `clean` | `rm -rf ./bin` |
| `fmt` | `gofmt -w .` |
| `deps` | `go mod tidy` |

**GoReleaser (.goreleaser.yaml):**
- Version 2 schema
- Pre-build hooks: `go mod tidy`, `go generate ./...`
- CGO_ENABLED=0
- Targets: Linux, Windows, Darwin (macOS)
- Archives: tar.gz for Linux/macOS, zip for Windows
- Name template: `<project>_<OS>_<arch>` with uname-compatible naming (amd64 -> x86_64, 386 -> i386)
- Changelog: sorted ascending, excludes `docs:` and `test:` prefixed commits
- Release footer links to GoReleaser

### Plugin Distribution

The project also includes `plugin.yaml` and `install-binary.sh` for distribution as a Helm plugin itself, allowing installation via `helm plugin install`.
