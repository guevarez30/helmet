# Helmet -> K9s UX Transformation: Gap Analysis & Plan

## Current State (Helmet)

Based on a full audit of the codebase (`ui.go`, all view/keymap files, `styles/`, `components/`):

- **Tab-based navigation** with `[` / `]` cycling through four fixed tabs: Releases, Repositories, Hub, Plugins
- **No command bar** -- navigation is strictly tab-based; there is no `:` prompt or any way to jump directly to a view
- **No search/filter on tables** -- the Hub view has a `/` search bar, but it queries the Artifact Hub API; there is no client-side filtering of any table rows
- **No keyboard-driven resource jumping** -- users must cycle tabs sequentially with `[`/`]`; no way to go directly to a specific view by name
- **Basic keybindings per view** -- `i` (install), `D` (delete), `u` (upgrade/update), `r` (refresh), `R` (rollback in history), `enter`/`space` (select/drill), `esc` (back), `v` (show default values in Hub/Repos)
- **Fixed tab structure** -- four hardcoded tabs, no dynamic view creation or custom layouts
- **No breadcrumbs or status bar** -- the top bar is a horizontal tab menu rendered via `renderMenu()`; the bottom bar is a `help.Model` showing keybinding hints inline
- **Purple highlight theme, rounded borders** -- `HighlightColor` is `#874BFD`/`#7D56F4`, all borders use `lipgloss.RoundedBorder()`, all colors are hardcoded in `styles/styles.go`
- **Partial vim-style navigation** -- Repositories view supports `h`/`j`/`k`/`l` for panel and row movement; Releases detail view supports `h`/`l` for sub-tab switching; but `j`/`k` are NOT bound in Releases overview (only arrow keys via the bubbles table default keymap); `g`/`G` are explicitly **unbound** in `GenerateTable()` (`k.GotoBottom.Unbind()`, `k.GotoTop.Unbind()`)
- **No sorting capability** -- tables have fixed column order, no sort-by-column mechanism
- **Help shown as bottom bar, not toggleable panel** -- `help.Model` renders a single-line `ShortHelp()` at the bottom; `FullHelp()` returns empty slices everywhere, so `?` toggle does nothing
- **No wide/compact view modes** -- columns use a fixed/flex system (`components.ColumnDefinition`) but no toggle to show/hide optional columns
- **No multi-select** -- single row selection only across all tables
- **No confirmation dialogs** except for release deletion (`y`/`n` prompt)
- **Viewport-based detail views** -- Notes, Metadata, Hooks, Values, Manifest each use `viewport.Model` for scrollable text
- **Working directory** -- creates `~/.helm-tui/` on startup for workspace storage
- **Editor integration** -- has `helpers/editor.go` for opening `$EDITOR` (values editing in install/upgrade flows)

## Target State (K9s-like)

The target is a k9s-inspired terminal UI that preserves Helmet's Helm-specific domain while adopting k9s interaction patterns:

- **Command bar (`:`)** for jumping between views by name with autocomplete
- **Filter bar (`/`)** with regex, inverse (`/!`), and fuzzy (`/-f`) modes on all table views
- **Vim-style navigation** (`j`/`k`/`h`/`l`/`g`/`G`/`Ctrl-f`/`Ctrl-b`) everywhere
- **Breadcrumb header** showing current navigation context (e.g., `Helmet > Releases > nginx > Values`)
- **Status bar** at bottom showing filter state, item count, last action feedback
- **Context-sensitive help panel** (`?`) as a toggleable overlay listing all keybindings for the current view
- **Column sorting** (`Shift+letter`) with visual sort indicator
- **Multi-select** (`Space`) for batch operations
- **Theming/skin system** -- YAML-based skin files for customizable colors
- **Plugin system** for extensibility with custom commands and keybindings
- **Wide/compact view toggle** to show/hide optional columns
- **Describe/YAML/Log view equivalents** for Helm resources -- already partially present as viewport-based detail views

---

## Gap Analysis

### Navigation Gaps

| K9s Feature | Helmet Current | Gap | Severity |
|---|---|---|---|
| `:` command bar | None -- tab cycling only via `[`/`]` | **MISSING** -- Need full command bar with autocomplete, mapped to resource views | HIGH |
| `/` filter bar (client-side) | Only Hub has `/` search, but it calls the Artifact Hub API (server-side), not client-side filter | **MISSING** -- Need client-side regex filter on ALL table views | HIGH |
| Breadcrumbs | Top bar is a tab menu (`renderMenu()` in `ui.go`), not a context path | **MISSING** -- Need hierarchical path display (e.g., `Releases > nginx > History`) | MEDIUM |
| `[`/`]` history navigation | Used for sequential tab cycling (wraps around) | **DIFFERENT** -- K9s uses `[`/`]` for back/forward in view history; Helmet uses them for tab switching | MEDIUM |
| Esc back navigation | Per-view: resets `selectedView` to parent (e.g., detail -> releases list), but no unified stack | **PARTIAL** -- Need a consistent back-stack across ALL views, not per-model ad hoc logic | MEDIUM |
| Vim `j`/`k` row movement | Repos view binds `j`/`k` (via `"down"`, `"up"`, `"j"`, `"k"` in keyMsg switch); Releases overview does NOT -- relies on bubbles table default (arrow keys only) | **INCONSISTENT** -- Some views have it, some do not; need uniform binding | HIGH |
| Vim `g`/`G` (top/bottom) | **Explicitly unbound** in `GenerateTable()`: `k.GotoBottom.Unbind()`, `k.GotoTop.Unbind()` | **MISSING** -- Intentionally removed; must re-enable | HIGH |
| `Ctrl-f`/`Ctrl-b` page scroll | **Explicitly unbound** in `GenerateTable()`: `k.HalfPageUp.Unbind()`, `k.PageDown.Unbind()`, `k.HalfPageDown.Unbind()` | **MISSING** -- Intentionally removed; must re-enable | HIGH |
| Autocomplete in command bar | Install wizard uses `textinput` with suggestions (Repos install has `SuggestionKeyMap`); no command bar exists | **PARTIAL** -- Suggestion infrastructure exists but needs new command bar context | MEDIUM |

### Search & Filter Gaps

| K9s Feature | Helmet Current | Gap | Severity |
|---|---|---|---|
| Regex filter `/` on tables | No client-side filtering on any table; Hub `/` triggers API search | **MISSING** -- Must add per-table row filtering with regex support | HIGH |
| Inverse filter `/!` | None | **MISSING** -- Exclude-pattern mode | MEDIUM |
| Fuzzy filter `/-f` | None | **MISSING** -- Fuzzy/substring matching mode | LOW |
| Label filter `/-l` | None | **N/A** -- Helm resources do not have Kubernetes labels in the same sense; could adapt to namespace or status filter | N/A |
| Live filtering (type-as-you-go) | None -- Hub search requires Enter to submit | **MISSING** -- Filter should update rows as user types | MEDIUM |
| Sort by column | No sort capability anywhere; columns are fixed order | **MISSING** -- Need clickable/keybind-driven column sort with direction indicator | HIGH |
| Search result highlighting | None | **MISSING** -- Matched text should be visually highlighted in filtered rows | LOW |

### Visual/Layout Gaps

| K9s Feature | Helmet Current | Gap | Severity |
|---|---|---|---|
| Header with cluster/context info | None -- no header area showing kubeconfig context or namespace | **MISSING** -- Need top bar showing active kube context and optional namespace filter | MEDIUM |
| Breadcrumb bar | Tab bar (horizontal `renderMenu()` showing Releases/Repositories/Hub/Plugins) -- different concept | **MISSING** -- Need hierarchical path that updates as user drills into views | MEDIUM |
| Status bar (bottom) | Bottom area is `help.Model.View()` rendering `ShortHelp()` keybindings inline | **PARTIAL** -- Can repurpose bottom bar to show: filter state, item count, last action, key hints | MEDIUM |
| Color-coded status | No color differentiation -- all table rows use same style; `Selected` style is purple bg (color `57`) with light fg (color `229`) | **MISSING** -- Releases should be color-coded by status (deployed=green, failed=red, pending=yellow, etc.) | HIGH |
| Wide/compact toggle | Fixed columns via `ColumnDefinition` with `Width` or `FlexFactor`; no concept of optional/hidden columns | **MISSING** -- Need column visibility toggle | LOW |
| Fullscreen resource view | Viewport-based detail views fill remaining space but no true fullscreen toggle | **MISSING** -- `Ctrl+f` or similar to maximize current pane | LOW |
| Toggleable header | No -- tab menu always visible, takes space from content area (`msg.Height - lipgloss.Height(m.renderMenu())`) | **MISSING** -- `Ctrl+e` to hide/show header and reclaim vertical space | LOW |
| Skin/theme system | Single hardcoded theme in `styles/styles.go` -- `HighlightColor` is `#874BFD`/`#7D56F4`, selection is colors `57`/`229` | **MISSING** -- Need external theme file support (YAML or TOML) | LOW |
| Logo/splash display | None -- shows "loading..." during init | **MISSING** -- Could show Helmet logo briefly on startup like k9s does | LOW |

### Action/Interaction Gaps

| K9s Feature | Helmet Current | Gap | Severity |
|---|---|---|---|
| `d` describe resource | Metadata viewport shows parsed release metadata (`metadataVP`) -- partial equivalent | **PARTIAL** -- Exists as a sub-tab, but not triggered by single `d` keypress; requires Enter then `l`/`h` to navigate to Metadata tab | MEDIUM |
| `y` YAML view | Manifest viewport (`manifestVP`) shows Helm manifest output | **EXISTS** -- Different access pattern (sub-tab navigation) but functional equivalent present | LOW |
| `e` edit in $EDITOR | Values editing exists in install and upgrade flows via `helpers/editor.go` and `types.EditorFinishedMsg` | **PARTIAL** -- Only for install/upgrade value editing; no general "edit resource" from any view | MEDIUM |
| Multi-select (`Space`) | None -- `enter`/`space` is bound to "select/drill into" in Releases | **MISSING** -- Need row multi-select for batch operations; `Space` currently conflicts with drill-down | HIGH |
| Confirmation dialogs | Only release delete has `y`/`n` confirmation; repo delete (`D` in Repos) executes immediately | **PARTIAL** -- Inconsistent: some destructive actions confirm, others do not | MEDIUM |
| `Ctrl+s` screendump | None | **MISSING** -- Terminal screenshot/dump capability | LOW |
| `?` help panel (overlay) | `help.Model` renders `ShortHelp()` inline at bottom; `FullHelp()` returns empty arrays in ALL keymap implementations | **DIFFERENT** -- Infrastructure exists (bubbles `help.Model` supports full help) but `FullHelp()` is deliberately empty everywhere; need to populate and make it a toggleable overlay | MEDIUM |
| `Ctrl+d` delete resource | `D` (shift+d) deletes in both Releases and Repos | **DIFFERENT** -- Different keybinding but equivalent functionality | LOW |
| `Ctrl+k` kill/force delete | None | **MISSING** -- Could map to `helm uninstall --no-hooks` or similar forced removal | LOW |
| `Ctrl+z` suspend to shell | None | **MISSING** -- Bubbletea supports `tea.Suspend` but not wired up | LOW |

### Missing Helm-Specific Features (Beyond K9s Analogy)

These are capabilities that k9s provides for Kubernetes that have meaningful Helm equivalents helmet should implement:

1. **Namespace filtering** -- Currently all releases are listed from all namespaces in a flat table. Need: namespace filter in command bar (`:releases -n kube-system`), quick namespace switch, "all namespaces" vs single namespace mode. K9s equivalent: namespace dropdown and `:ns` command.

2. **Status color coding** -- Helm release statuses (deployed, failed, pending-install, pending-upgrade, pending-rollback, superseded, uninstalling, uninstalled) have clear semantic meaning that should map to colors: deployed=green, failed=red, pending-*=yellow, superseded=gray, uninstalling=orange. Currently all rows render identically.

3. **Real-time refresh** -- The Releases view only refreshes on explicit `r` keypress. Need: configurable auto-refresh interval (e.g., every 5s), visual staleness indicator, `Ctrl+r` manual refresh. K9s equivalent: configurable refresh rate in config.

4. **Diff view** -- Compare values between two revisions of a release. Show side-by-side or unified diff of Helm values. Accessible from history view when two revisions are available. K9s equivalent: diff between resource versions.

5. **Dependency tree** -- Show chart dependency hierarchy (parent chart -> subcharts). Helm charts can have complex dependency trees. K9s equivalent: XRay tree view for resource relationships.

6. **Bulk operations** -- Multi-select releases for batch delete or batch rollback. Currently only single-resource operations. K9s equivalent: `Space` multi-select + batch action.

7. **Release notes diff** -- Show what changed between chart versions when upgrading, sourced from CHANGELOG or chart annotations.

8. **Resource mapping** -- Show which Kubernetes resources a Helm release created (pods, services, configmaps, etc.). K9s equivalent: ownerRef-based resource tree.

---

## Phased Implementation Plan

### Phase 1: Core Navigation Overhaul (Foundation)

**Goal**: Replace tab-based navigation with k9s-style command-driven navigation. This is the foundation that all subsequent phases build upon.

**Estimated scope**: ~15-20 files modified/created, significant refactor of `ui.go` and the main model's Update/View loop.

#### 1.1 Command Bar

**What**: Add a persistent `textinput.Model` at the bottom of the screen, activated by pressing `:`.

**Implementation details**:
- Add a `commandBar textinput.Model` field to `mainModel` in `ui.go`
- Add a `commandBarActive bool` field to track activation state
- When `:` is pressed and command bar is not active, focus the command bar input
- Implement resource type commands: `:releases`, `:repos`, `:hub`, `:plugins`
- Add shorthand aliases: `:rel`, `:repo`, `:h`, `:plug`
- Wire up `tab` key to cycle through autocomplete suggestions using existing `SuggestionKeyMap` infrastructure from `helpers/keymaps.go`
- `Enter` executes the command (switches to target view)
- `Escape` dismisses the command bar without action
- Render the command bar at the bottom of the screen in `View()`, replacing or sitting above the help bar
- The command bar replaces `[`/`]` tab cycling as the **primary** navigation method (but `[`/`]` remain as shortcuts)

**Files affected**: `ui.go` (main model, Update, View), `helpers/keymaps.go` (add command bar keys)

**Dependencies**: None -- this is the foundation.

#### 1.2 Vim-Style Movement

**What**: Enable consistent vim movement keybindings across all table views.

**Implementation details**:
- In `components/table.go` `GenerateTable()`, **stop unbinding** `GotoTop`, `GotoBottom`, `HalfPageUp`, `HalfPageDown`, `PageDown`:
  - Remove: `k.HalfPageUp.Unbind()`, `k.PageDown.Unbind()`, `k.HalfPageDown.Unbind()`, `k.GotoBottom.Unbind()`, `k.GotoTop.Unbind()`
  - The bubbles `table` component already maps `g`/`G` to GotoTop/GotoBottom and `Ctrl-f`/`Ctrl-b` to page movement when these bindings are active
- Verify `j`/`k` work in ALL views -- currently missing in Releases overview (the bubbles table default binds up/down arrows but not j/k; need to add j/k to the table keymap or handle in the Update function)
- Add `h`/`l` for lateral panel navigation consistently (already present in Repos and Releases detail, but formalize)
- Ensure vim bindings are suppressed when any `textinput` is focused (command bar, search bar, install wizard inputs) to avoid conflicts

**Files affected**: `components/table.go`, `releases/overview.go`, `hub/hub.go`, `plugins/overview.go`

**Dependencies**: None.

#### 1.3 Breadcrumb Header

**What**: Add a header bar showing the current navigation context path.

**Implementation details**:
- Add a `breadcrumbs []string` field to `mainModel` that tracks the navigation path
- Render format: `Helmet > Releases` at top level, `Helmet > Releases > nginx-release > Values` when drilling into detail
- Each view module must emit a message or expose a method to report its current breadcrumb segment
- Add a new `types.BreadcrumbMsg` message type to communicate breadcrumb updates from child models to the main model
- Replace the current tab menu (`renderMenu()`) with a breadcrumb bar that also shows available top-level views (dimmed, with the active one highlighted)
- Make the header toggleable with `Ctrl+e` -- when hidden, reclaim the vertical space for content
- Consider: show the active kubeconfig context name on the far right of the header (from `$KUBECONFIG` or `~/.kube/config` parsing)

**Files affected**: `ui.go` (renderMenu replacement), `types/messages.go` (new message type), all view modules (emit breadcrumb messages)

**Dependencies**: 1.1 (command bar informs breadcrumb state).

#### 1.4 Status Bar

**What**: Replace the bottom help bar with a k9s-style status bar.

**Implementation details**:
- The current bottom bar is rendered per-view via `m.help.View(m.keys[m.selectedView])` (see `releases/overview_view.go` line 38)
- Replace with a unified status bar component rendered by `mainModel.View()` (not by individual views)
- Status bar layout: `[left: current filter expression] [center: item count "42 releases"] [right: context-sensitive key hints]`
- Add a `types.StatusMsg` for child views to report item counts and action results (e.g., "Release nginx deleted", "3 repos updated")
- Flash messages: action results display for 3 seconds then clear
- Key hints: show the 4-5 most relevant keys for the current context (like k9s top-right hints, but at the bottom)

**Files affected**: `ui.go` (new status bar renderer), `types/messages.go`, all view modules (remove per-view help rendering, emit status messages)

**Dependencies**: 1.1 (status bar sits adjacent to command bar).

#### 1.5 Back-Stack Navigation

**What**: Implement a proper view history stack so navigation is predictable.

**Implementation details**:
- Add a `viewStack []viewState` to `mainModel` where `viewState` captures: tab index, selected view within the tab, selected row cursor position
- `Enter`/drill-down pushes the current state onto the stack
- `Escape` pops the last state and restores it (go back one level)
- `-` key returns to the previous view (like vim's `-` or k9s back shortcut) -- this is different from `Escape` in that it can cross tab boundaries
- Reclaim `[`/`]` from tab cycling to instead mean back/forward in history (matching k9s convention); tab switching moves entirely to the command bar
- Maximum stack depth of 50 to prevent memory issues
- Current per-view back logic in `releases/overview.go` (the `esc` handler that resets `selectedView` and restores `releaseTableCache`) becomes a special case of the general back-stack

**Files affected**: `ui.go` (stack management), all view modules (remove per-view back navigation logic, delegate to main model)

**Dependencies**: 1.1, 1.3 (breadcrumbs update from stack state).

---

### Phase 2: Search & Filter System

**Goal**: Add k9s-style client-side filtering to all table views so users can narrow down large lists instantly.

**Estimated scope**: ~8-12 files modified/created, new shared filter component.

#### 2.1 Global Filter Bar

**What**: `/` activates a filter input on any table view; typed text filters table rows in real time.

**Implementation details**:
- Create a new shared component `components/filter.go` containing a `FilterModel` wrapping `textinput.Model`
- When `/` is pressed (and no other input is focused), activate the filter bar above the status bar
- As the user types, filter the table's `Rows()` against the filter pattern using Go's `regexp` package
- Matching is case-insensitive by default
- Filter applies across ALL visible columns of the table (any column match keeps the row)
- `Enter` confirms the filter and returns focus to the table (filter remains active)
- `Escape` clears the filter and restores full row set
- Visual indicator: show the active filter expression in the status bar (Phase 1.4) and in the filter input itself
- Each view model that contains a `table.Model` gains a `FilterModel` and stores the original unfiltered rows separately

**Files affected**: New `components/filter.go`, `releases/overview.go`, `repositories/overview.go`, `hub/hub.go`, `plugins/overview.go`

**Dependencies**: Phase 1.4 (filter state shown in status bar).

#### 2.2 Filter Modes

**What**: Support multiple filter modes beyond plain regex.

**Implementation details**:
- `/pattern` -- standard regex filter (default mode, implemented in 2.1)
- `/!pattern` -- inverse regex: show rows that do NOT match the pattern; detect the `!` prefix and invert the match logic
- `/-f pattern` -- fuzzy matching: detect the `-f ` prefix and use a fuzzy matching algorithm (e.g., subsequence match with scoring) instead of regex
- Mode indicator: show the active mode in the filter bar (e.g., `[regex]`, `[inverse]`, `[fuzzy]`)
- Implement mode detection in `FilterModel.Update()` by parsing the input prefix before applying the filter function
- Consider adding `/-s column:pattern` for column-specific filtering in the future (not in initial implementation)

**Files affected**: `components/filter.go` (mode parsing and matching logic)

**Dependencies**: 2.1 (base filter infrastructure).

#### 2.3 Column Sorting

**What**: Allow sorting table rows by any column via keyboard shortcuts.

**Implementation details**:
- Shift+key triggers sort by a specific column:
  - **Releases**: `Shift+N` name, `Shift+S` status, `Shift+A` age/updated, `Shift+R` revision, `Shift+C` chart, `Shift+V` app version
  - **Repositories**: `Shift+N` name, `Shift+U` URL
  - **Hub results**: `Shift+P` package, `Shift+R` repository
  - **Plugins**: `Shift+N` name, `Shift+V` version
- Pressing the same sort key again toggles ascending/descending
- Visual indicator: show an arrow (up/down unicode char) next to the sorted column header title
- Implement sort in `components/table.go` as a helper that takes `[]table.Row`, column index, direction, and returns sorted rows
- Sort interacts with filter: sort applies to the filtered result set
- Default sort: by name ascending (matches current implicit order from `helm list` output)

**Files affected**: `components/table.go` (sort helper), all view keymap files (add Shift+key bindings), all view models (wire sort state)

**Dependencies**: 2.1 (sort applies on top of filter).

---

### Phase 3: Visual Polish (K9s Look & Feel)

**Goal**: Bring Helmet's visual aesthetic closer to k9s with status colors, refined layout, and a proper help system.

**Estimated scope**: ~10-15 files modified, primarily styling and view rendering changes.

#### 3.1 Status Color Coding

**What**: Color-code table rows based on semantic status.

**Implementation details**:
- Define a color palette for Helm release statuses in `styles/styles.go`:
  - `deployed` = green (`#00FF00` or lipgloss adaptive)
  - `failed` = red (`#FF0000`)
  - `pending-install`, `pending-upgrade`, `pending-rollback` = yellow (`#FFFF00`)
  - `superseded` = gray (`#808080`)
  - `uninstalling` = orange (`#FFA500`)
  - `uninstalled` = dim gray (`#404040`)
- Modify the Releases table rendering to apply row-level styling based on the Status column value
- The bubbles `table` component renders rows via its `Styles` field; we may need to post-process the rendered view string or use a custom row renderer
- Alternative approach: since bubbles table does not natively support per-row colors, render a custom table component or override the `View()` output with lipgloss styling per-line
- Repos: color repo entries by update status (if available)
- Hub: color by verification status or star count tier
- Plugins: color by installed vs available for update

**Files affected**: `styles/styles.go` (new color definitions), `releases/overview_view.go`, `repositories/overview_view.go`, `hub/hub_view.go`, `plugins/overview_view.go`

**Dependencies**: None -- can be done in parallel with Phase 2.

#### 3.2 Layout Refinement

**What**: Restructure the screen layout to match k9s conventions.

**Implementation details**:
- **Header area** (top): Left side shows Helmet logo/name + breadcrumb path; right side shows kubeconfig context name, namespace (if filtered), and 4-5 key hints
- Parse kubeconfig context from environment or `~/.kube/config` to display in header
- **Content area** (middle): Table or viewport, unchanged
- **Footer area** (bottom): Status bar (left: filter, center: counts, right: action flash) + command bar (shown only when active)
- Remove the current tab menu entirely -- it is replaced by breadcrumbs in the header
- Clean up table borders: consider using thinner borders or removing side borders to maximize horizontal space (k9s uses minimal borders)
- Evaluate switching from `lipgloss.RoundedBorder()` to `lipgloss.NormalBorder()` or a custom minimal border for a cleaner look

**Files affected**: `ui.go` (new layout structure), `styles/styles.go` (border adjustments), all `*_view.go` files (layout constants)

**Dependencies**: Phase 1.3 (breadcrumbs), Phase 1.4 (status bar).

#### 3.3 Help Panel

**What**: `?` toggles a full help panel overlay showing all keybindings for the current view.

**Implementation details**:
- Populate `FullHelp()` in ALL keymap structs (currently ALL return empty `[][]key.Binding{}`):
  - `helpers/keymaps.go` -- common keys
  - `releases/overview_keymap.go` -- release-specific keys grouped by category
  - `repositories/overview_keymap.go` -- repo-specific keys
  - `hub/hub_keymap.go` -- hub-specific keys
  - `plugins/overview_keymap.go` -- plugin-specific keys
- Create a help overlay component that renders `FullHelp()` output in a centered, bordered panel
- `?` key toggles the overlay on/off
- When overlay is active, all other keys are suppressed except `?` (toggle off) and `Escape` (dismiss)
- The overlay should show: view name at top, keybindings grouped by category (Navigation, Actions, Filters, etc.)
- Overlay should respect terminal size and scroll if keybindings exceed available height

**Files affected**: All `*_keymap.go` files (populate FullHelp), new `components/helpoverlay.go`, `ui.go` (overlay toggle logic)

**Dependencies**: None -- can be done independently.

#### 3.4 Wide/Compact View Toggle

**What**: Toggle between compact (default) and wide (all columns) table views.

**Implementation details**:
- `Ctrl+w` toggles between compact and wide mode
- Define two column sets per view:
  - **Releases compact**: Name, Namespace, Status, Chart, App Version (5 columns)
  - **Releases wide**: Name, Namespace, Revision, Updated, Status, Chart, App Version (7 columns -- current default is already wide)
  - **Repos compact**: Name only
  - **Repos wide**: Name, URL
  - Adjust similarly for Hub and Plugins
- Store a `wideMode bool` in each view model
- When toggled, call `components.SetTable()` with the appropriate column definition set
- Show current mode indicator in status bar: `[wide]` or `[compact]`

**Files affected**: All view models (add `wideMode` field and column sets), all `*_view.go` files (conditional column selection)

**Dependencies**: Phase 1.4 (mode indicator in status bar).

---

### Phase 4: Enhanced Helm Operations

**Goal**: Add Helm-specific features that exploit the domain knowledge helmet has, creating value beyond what k9s offers for Helm.

**Estimated scope**: ~15-20 files modified/created, significant new Helm command integrations.

#### 4.1 Namespace Filtering

**What**: Filter releases by Kubernetes namespace rather than always showing all namespaces.

**Implementation details**:
- Add a `:ns` command in the command bar to switch namespace context
- `:ns kube-system` filters to a single namespace; `:ns all` shows all (default)
- Add `Shift+N` as a quick namespace filter shortcut (if not conflicting with sort -- use a different key if needed)
- Store `activeNamespace string` in the releases model
- Pass `--namespace` flag to `helm list` commands when a namespace is set
- Show active namespace in the header bar (Phase 3.2): `Helmet > Releases [ns: kube-system]`
- When namespace is set, the status bar item count reflects filtered count
- Implement namespace autocomplete in the command bar by listing available namespaces from `kubectl get ns`

**Files affected**: `releases/overview.go` (namespace state), `releases/overview_commands.go` (helm command modification), `ui.go` (command bar namespace handling)

**Dependencies**: Phase 1.1 (command bar), Phase 1.3 (breadcrumb namespace display).

#### 4.2 Diff View

**What**: Compare Helm values between two revisions of a release.

**Implementation details**:
- Accessible from the History sub-view of a release
- New keybinding: `d` when in history view to "diff this revision against the previous one"
- Alternative: select two revisions with multi-select (Phase prerequisite) and diff between them
- Run `helm get values RELEASE --revision N` for both revisions and compute a unified diff
- Display diff in a new viewport with syntax highlighting:
  - Added lines in green with `+` prefix
  - Removed lines in red with `-` prefix
  - Context lines in default color
- Use Go's `text/template` or a diff library for computing the diff (consider `github.com/sergi/go-diff`)
- Add `Diff` as a new sub-tab alongside History, Notes, Metadata, Hooks, Values, Manifest

**Files affected**: `releases/overview.go` (new `diffView` state), new `releases/diff.go` (diff computation), `releases/overview_view.go` (render diff), `releases/overview_keymap.go` (new `d` binding)

**Dependencies**: Phase 1 (navigation), ideally Phase 2 multi-select for two-revision selection.

#### 4.3 Real-Time Refresh

**What**: Automatically refresh data on a configurable interval.

**Implementation details**:
- Add a `tea.Tick` command that fires every N seconds (default: 5s, configurable)
- On each tick, re-fetch the data for the currently active view only (avoid unnecessary API calls)
- Show a visual freshness indicator in the status bar: timestamp of last refresh, or a spinner during refresh
- `Ctrl+r` forces an immediate manual refresh (supplement existing `r` key which already refreshes in some views)
- Add a `--refresh-rate` CLI flag and config file option
- Stale data indicator: if refresh fails (helm command error), show a warning in the status bar
- Pause auto-refresh when the user is in the middle of an action (installing, upgrading, deleting, typing in command/filter bar)

**Files affected**: `ui.go` (tick management), `main.go` (CLI flag), all view commands files (refresh coordination), config system

**Dependencies**: Phase 1.4 (status bar for refresh indicator).

#### 4.4 Describe View

**What**: A comprehensive "describe" output for a release, consolidating multiple data sources into a single scrollable view.

**Implementation details**:
- `d` key on a selected release shows a describe panel combining:
  - Release name, namespace, status, chart, app version, revision number
  - First deployed / last deployed timestamps
  - Revision history summary (last 5 revisions, one line each)
  - Chart metadata (maintainers, home URL, sources, description)
  - Notes (abbreviated)
- This is distinct from the existing individual viewports (Metadata, Notes, etc.) -- it is a unified summary
- Render as a scrollable viewport with section headers
- Accessible directly from the releases list view (no need to Enter first)
- K9s analogy: `d` on a pod shows `kubectl describe pod`; this is `helm describe release` (which doesn't exist as a single command, so we synthesize it)

**Files affected**: New `releases/describe.go` (data fetching and formatting), `releases/overview.go` (new view state), `releases/overview_view.go` (render), `releases/overview_keymap.go` (new `d` binding)

**Dependencies**: None.

#### 4.5 YAML View

**What**: Show raw YAML/JSON of the release with syntax highlighting.

**Implementation details**:
- `y` key on a selected release shows the full release object as YAML
- Use `helm get all RELEASE` or combine multiple `helm get` outputs
- Apply syntax highlighting using ANSI color codes:
  - Keys in cyan
  - String values in green
  - Numbers in yellow
  - Booleans in magenta
  - Comments in gray
- Consider using `github.com/alecthomas/chroma` for syntax highlighting or a lightweight custom YAML colorizer
- Render in a viewport with line numbers
- `Ctrl+c` in YAML view copies the content to clipboard (if terminal supports OSC 52)

**Files affected**: `releases/overview.go` (new `yamlView` state), `releases/overview_view.go` (render with highlighting), `releases/overview_keymap.go` (`y` binding at top level)

**Dependencies**: None.

---

### Phase 5: Extensibility & Theming

**Goal**: Enable users to customize helmet's appearance and extend its functionality, following k9s's proven extensibility model.

**Estimated scope**: ~10-15 new files, new configuration system.

#### 5.1 Theme/Skin System

**What**: Allow users to customize all UI colors via external configuration files.

**Implementation details**:
- Create a `~/.helm-tui/skins/` directory for theme files
- Theme file format: YAML (matching k9s convention), e.g.:
  ```yaml
  skin:
    background: "#1a1b26"
    foreground: "#c0caf5"
    highlight: "#7aa2f7"
    border: "#3b4261"
    status:
      deployed: "#9ece6a"
      failed: "#f7768e"
      pending: "#e0af68"
      superseded: "#565f89"
    table:
      header_fg: "#7aa2f7"
      header_bg: "#1a1b26"
      selected_fg: "#1a1b26"
      selected_bg: "#7aa2f7"
      cursor_fg: "#c0caf5"
  ```
- Refactor `styles/styles.go` to load colors from the active skin file instead of hardcoding
- Add `--skin` CLI flag to select a skin by name
- Ship with built-in themes: `default` (current purple), `tokyo-night`, `catppuccin`, `dracula`, `solarized-dark`, `solarized-light`
- Skin hot-reload: detect file changes and re-apply (stretch goal)
- Per-context theming: different skin per kubeconfig context (like k9s), so production clusters can be visually distinct (e.g., red borders for prod)

**Files affected**: `styles/styles.go` (complete refactor to dynamic colors), new `config/skin.go` (skin loader), `main.go` (CLI flag)

**Dependencies**: Phase 3.1 (status color coding defines the color semantics the skin system must support).

#### 5.2 Plugin System

**What**: Allow users to define custom actions that integrate external commands.

**Implementation details**:
- Plugin definition in `~/.helm-tui/plugins.yaml`:
  ```yaml
  plugins:
    log-pods:
      shortCut: Ctrl-L
      description: "View pods for release"
      command: kubectl
      args: ["get", "pods", "-l", "app.kubernetes.io/instance=$NAME", "-n", "$NAMESPACE"]
      background: false
    open-browser:
      shortCut: Ctrl-O
      description: "Open chart homepage"
      command: open
      args: ["$CHART_HOME"]
      background: true
  ```
- Context variables resolved at execution time: `$NAME`, `$NAMESPACE`, `$CHART`, `$VERSION`, `$REVISION`, `$STATUS`, `$REPO_URL`
- Plugins appear in the help panel (`?`) under a "Plugins" section
- Plugin output shown in a new viewport (for non-background plugins) or suppressed (for background plugins)
- Plugin execution uses `tea.ExecProcess` (like the existing editor integration) for foreground commands
- Validate plugin definitions on startup and warn about conflicts with built-in keybindings

**Files affected**: New `config/plugins.go` (plugin loader and executor), `ui.go` (plugin keybinding registration), `types/messages.go` (plugin result message), help panel (show plugin bindings)

**Dependencies**: Phase 3.3 (help panel to display plugin keybindings), Phase 1.1 (command bar for `:plugins` command).

#### 5.3 Configuration File

**What**: Centralized application configuration for all settings.

**Implementation details**:
- Config file: `~/.helm-tui/config.yaml`
  ```yaml
  helmet:
    refreshRate: 5          # seconds, 0 = disabled
    defaultView: releases   # initial view on startup
    defaultNamespace: all   # or specific namespace
    ui:
      skin: default
      wideMode: false
      showHeader: true
      showStatusBar: true
    keybindings:
      commandBar: ":"
      filterBar: "/"
      back: "esc"
      help: "?"
    helm:
      binary: helm           # path to helm binary
      kubeconfig: ""         # override kubeconfig path
  ```
- Parse on startup with `gopkg.in/yaml.v3` or `github.com/spf13/viper`
- CLI flags override config file values; config file overrides defaults
- Key rebinding: allow users to remap any built-in keybinding (with validation for conflicts)
- Config reload: `:config reload` command in the command bar to re-read config without restarting
- First-run experience: if no config exists, use sensible defaults (current behavior)

**Files affected**: New `config/config.go` (config loader), `main.go` (config initialization, CLI flag merging), `ui.go` (apply config values), all view models (read config for defaults)

**Dependencies**: Phase 5.1 (skin reference in config), Phase 5.2 (plugin file reference in config).

---

### Phase 6: UltraSearch (Novel Feature -- Beyond K9s)

**Goal**: A cross-cutting search capability that goes beyond anything k9s offers, leveraging the unique position of helmet as a Helm-specific tool that can search across installed releases, configured repositories, and the public Artifact Hub simultaneously.

**Estimated scope**: ~8-12 new/modified files, new search infrastructure.

#### 6.1 Cross-View Search

**What**: A single search that queries ALL views simultaneously and presents unified results.

**Implementation details**:
- Activate via `Ctrl+/` or `:search` command
- Opens a full-screen search overlay with a text input at the top
- As the user types, simultaneously search:
  - Installed releases (name, chart, namespace, status)
  - Repository packages (name, description)
  - Artifact Hub results (name, description, repository)
  - Installed plugins (name, description)
- Results grouped by source with section headers:
  ```
  --- Installed Releases (3 matches) ---
  nginx-ingress    default    deployed    ingress-nginx/ingress-nginx
  nginx-test       staging    failed      bitnami/nginx
  nginx-proxy      prod       deployed    custom/nginx-proxy

  --- Repository Charts (5 matches) ---
  bitnami/nginx             A chart for deploying nginx
  ingress-nginx/ingress-nginx    Ingress controller for Kubernetes

  --- Artifact Hub (12 matches) ---
  nginx (Bitnami)           Deploy nginx on Kubernetes
  nginx-ingress (NGINX Inc) F5 NGINX Ingress Controller
  ```
- Each result is navigable: `Enter` on a result jumps to that item in its native view
- Results update in real-time as the user types (debounced: 200ms delay before triggering Hub search to avoid API spam; local searches are instant)
- `Escape` dismisses the search overlay

**Files affected**: New `search/search.go` (search model and orchestration), new `search/search_view.go` (rendering), `ui.go` (overlay integration), `types/messages.go` (search result messages)

**Dependencies**: Phase 1.1 (command bar for `:search`), Phase 2.1 (filter infrastructure for matching logic).

#### 6.2 Smart Search

**What**: Context-aware search that understands Helm concepts and groups results meaningfully.

**Implementation details**:
- When the user searches a chart name (e.g., "nginx"), the results are organized by intent:
  - **Installed**: releases using this chart, with their status and version
  - **Available Upgrades**: newer versions available in configured repos for installed releases
  - **Install Options**: chart versions available in repos (not yet installed)
  - **Hub Discover**: matching results from Artifact Hub for charts not in any configured repo
- Each section offers one-action shortcuts:
  - On an installed release: `u` to upgrade, `D` to delete, `Enter` to view details
  - On an available upgrade: `Enter` to start upgrade flow
  - On an install option: `i` to start install flow
  - On a Hub result: `a` to add repo, `Enter` to view details
- Smart deduplication: if a chart exists both in a local repo and on Artifact Hub, show it once with both sources noted
- Chart version comparison: highlight when an installed release is behind the latest available version (e.g., "installed: 1.2.3, latest: 1.5.0" with the version delta in yellow)

**Files affected**: `search/search.go` (smart grouping logic), `search/search_view.go` (intent-based rendering), release/repo/hub command files (version comparison queries)

**Dependencies**: 6.1 (cross-view search infrastructure).

#### 6.3 Search History

**What**: Persist and navigate through previous searches.

**Implementation details**:
- Store last 100 searches in `~/.helm-tui/search_history`
- When the search overlay opens, show recent searches below the input (like a browser URL bar)
- Arrow keys navigate through history when the search input is focused
- `Ctrl+r` in the search overlay activates reverse search through history (like bash `Ctrl+r`)
- Bookmark/favorite frequently used searches: `Ctrl+b` bookmarks the current search
- Bookmarks shown in a separate section above history
- Bookmarks stored in `~/.helm-tui/bookmarks.yaml`
- `:history` command shows full search history in a navigable list

**Files affected**: New `search/history.go` (persistence and retrieval), `search/search.go` (history integration), `search/search_view.go` (history display)

**Dependencies**: 6.1 (search overlay), Phase 5.3 (config system for storage path).

---

## Priority Matrix

| Phase | Impact | Effort | Priority | Rationale |
|---|---|---|---|---|
| **Phase 1: Navigation** | HIGH | HIGH | **P0 -- Foundation** | Everything else depends on the command bar, vim navigation, and back-stack. Without this, subsequent phases cannot integrate properly. |
| **Phase 2: Search/Filter** | HIGH | MEDIUM | **P0 -- Core UX** | Filtering and sorting are the most-requested TUI features. A table without filter is painful at scale (50+ releases). |
| **Phase 3: Visual Polish** | MEDIUM | MEDIUM | **P1 -- Look & Feel** | Status colors and the help panel dramatically improve usability with moderate effort. Theming is lower priority but color coding is high. |
| **Phase 4: Helm Operations** | HIGH | HIGH | **P1 -- Differentiation** | Namespace filtering, diff view, and describe are what make helmet uniquely valuable vs. raw `helm` CLI. These are the features that justify using a TUI. |
| **Phase 5: Extensibility** | MEDIUM | HIGH | **P2 -- Long-term** | Theme and plugin systems create an ecosystem but require significant infrastructure. Worth doing after core UX is solid. |
| **Phase 6: UltraSearch** | HIGH | MEDIUM | **P2 -- Innovation** | Cross-view smart search is a genuinely novel feature no other Helm tool offers. It leapfrogs k9s rather than just copying it. Defer until the foundation is stable. |

### Recommended Execution Order

```
Phase 1.2 (Vim Movement)          -- Quick win, 1-2 days, unblocks developer flow
Phase 1.1 (Command Bar)           -- Core infrastructure, 3-5 days
Phase 1.5 (Back-Stack)            -- Depends on 1.1, 2-3 days
Phase 1.3 (Breadcrumbs)           -- Depends on 1.1/1.5, 2-3 days
Phase 1.4 (Status Bar)            -- Depends on 1.1, 2-3 days
  |
Phase 2.1 (Filter Bar)            -- Depends on 1.4, 3-5 days
Phase 2.3 (Column Sorting)        -- Depends on 2.1, 2-3 days
Phase 2.2 (Filter Modes)          -- Depends on 2.1, 1-2 days
  |
Phase 3.1 (Status Colors)         -- Independent, 2-3 days
Phase 3.3 (Help Panel)            -- Independent, 2-3 days
Phase 3.2 (Layout Refinement)     -- Depends on 1.3/1.4, 3-5 days
Phase 3.4 (Wide/Compact)          -- Depends on 1.4, 1-2 days
  |
Phase 4.1 (Namespace Filter)      -- Depends on 1.1, 3-5 days
Phase 4.3 (Real-Time Refresh)     -- Depends on 1.4, 2-3 days
Phase 4.4 (Describe View)         -- Independent, 3-5 days
Phase 4.5 (YAML View)             -- Independent, 2-3 days
Phase 4.2 (Diff View)             -- Depends on 4.4 patterns, 3-5 days
  |
Phase 5.3 (Config File)           -- Independent, 3-5 days
Phase 5.1 (Themes)                -- Depends on 5.3 + 3.1, 5-7 days
Phase 5.2 (Plugins)               -- Depends on 5.3 + 3.3, 5-7 days
  |
Phase 6.1 (Cross-View Search)     -- Depends on 2.1 + 1.1, 5-7 days
Phase 6.2 (Smart Search)          -- Depends on 6.1, 3-5 days
Phase 6.3 (Search History)        -- Depends on 6.1 + 5.3, 2-3 days
```

**Total estimated effort**: 60-90 developer days across all phases.

### Key Technical Risks

1. **Bubbles table limitations**: The `charmbracelet/bubbles/table` component does not natively support per-row coloring, client-side filtering, or column sort. We may need to fork it or build a custom table component. This could add 5-10 days to Phase 2/3.

2. **Keybinding conflicts**: Adding vim keys (`j`/`k`/`g`/`G`), filter (`/`), command bar (`:`), help (`?`), and sort (Shift+keys) to the existing keybinding space creates risk of collisions, especially when textinputs are focused. Need a clear keybinding state machine.

3. **Performance at scale**: Users with 100+ releases and 20+ repos -- filtering and sorting must remain responsive. All filtering should use compiled regex and pre-indexed data.

4. **Backward compatibility**: Users accustomed to current `[`/`]` tab navigation will need a transition period. Consider keeping `[`/`]` as tab shortcuts alongside the new command bar during Phase 1, then deprecating later.
