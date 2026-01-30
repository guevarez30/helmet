# Plan: Phase 1 — K9s Core Navigation Overhaul

## Context

Helmet (helm-tui) currently uses a simple tab-cycling UI (`[`/`]`). We are transforming it to match k9s interaction patterns: command bar, vim navigation, breadcrumbs, status bar, and a view back-stack. This is the foundation all future phases build on.

## Objectives

### Must Have
- Command bar (`:`) at bottom of screen with autocomplete for view switching
- Vim-style movement (j/k/g/G/Ctrl-f/Ctrl-b) in all table views
- Breadcrumb header replacing the current tab menu
- Status bar at bottom showing item count, filter state, flash messages
- View history back-stack with Escape-to-go-back

### Must NOT Have
- Changes to any Helm CLI commands or data flow
- New dependencies (use only existing bubbletea/bubbles/lipgloss)
- Removal of existing functionality (install, upgrade, delete wizards still work)
- Filter/search system (that's Phase 2)
- Theme/skin system (that's Phase 5)

## Pre-Work: Bug Fixes (Step 0)

Fix 3 bugs that would interfere with Phase 1 work.

### 0a. Fix WindowSizeMsg bug in ui.go

**File:** `ui.go:78-85`

**Bug:** The `cmd` variable is overwritten 4 times but only the last value (plugins) gets appended to `cmds`. The cmds slice is empty when `tea.Batch(cmds...)` runs.

**Fix:** Append each cmd after each Update call:
```go
case tea.WindowSizeMsg:
    m.width = msg.Width
    m.height = msg.Height
    adjustedMsg := tea.WindowSizeMsg{Width: m.width, Height: msg.Height - lipgloss.Height(m.renderMenu())}
    for i := range m.tabContent {
        m.tabContent[i], cmd = m.tabContent[i].Update(adjustedMsg)
        cmds = append(cmds, cmd)
    }
    return m, tea.Batch(cmds...)
```

### 0b. Fix hardcoded table title in components/table.go

**File:** `components/table.go:81-91`

**Bug:** `RenderTable()` hardcodes `" Releases "` as the title for all tabs.

**Fix:** Add `title` parameter:
```go
func RenderTable(t table.Model, height int, width int, title string) string {
    // ... use title instead of " Releases "
}
```

**Callers to update:**
- `releases/overview_view.go:97-105` — `renderReleasesTableView()` calls `RenderTable` (inlined, not via the function — but uses same pattern). Already uses `" Releases "` directly. Keep as-is since releases has its own render.
- `plugins/overview_view.go:17` — calls `components.RenderTable(...)`. Pass `" Plugins "`.

### 0c. Remove unused `index` field from mainModel

**File:** `ui.go:33`

**Fix:** Delete `index int` from the struct. No references exist.

---

## Step 1: Vim-Style Movement in All Tables

**Goal:** Enable j/k/g/G/Ctrl-f/Ctrl-b in every table view.

### 1a. Re-enable unbound keys in GenerateTable()

**File:** `components/table.go:56-78`

**Current code (lines 60-65):**
```go
k.HalfPageUp.Unbind()
k.PageDown.Unbind()
k.HalfPageDown.Unbind()
k.HalfPageDown.Unbind()  // duplicate
k.GotoBottom.Unbind()
k.GotoTop.Unbind()
```

**Fix:** Remove all 6 lines. The bubbles `table` default keymap already binds:
- `g` → GotoTop
- `G` → GotoBottom
- `ctrl+f` → PageDown
- `ctrl+b` → HalfPageUp
- `ctrl+d` → HalfPageDown

This immediately enables vim paging in ALL tables (releases, repos, hub, plugins) since they all use `GenerateTable()`.

### 1b. Verify j/k work everywhere

The bubbles `table.DefaultKeyMap()` already binds `up`/`down`/`k`/`j` for row movement. Since we're using the default keymap (minus our unbinds which we just removed), j/k should work everywhere.

**Verify in:** `releases/overview.go` — the KeyMsg handler at line 216 does NOT intercept `j`/`k`, so they fall through to `m.releaseTable.Update(msg)` at line 134 which handles them via the table's internal keymap. This is correct.

**Verify in:** `repositories/overview.go:173` — already handles `"down", "up", "j", "k"` for cascading refresh. The table Update at line 207 also processes them. Both work together correctly.

**No changes needed** — j/k already work via the table default keymap. We just need to not intercept them at the module level (which we don't, except repos which correctly handles the cascading side-effect).

### 1c. Handle g/G conflict with Releases

**Risk:** In `releases/overview.go`, the key handler doesn't intercept `g` or `G`. But the detail sub-tabs (history, notes, etc.) also pass through to their respective components. `g`/`G` on viewports is handled by the viewport's own keymap, which is fine.

**No changes needed** — no conflicts exist.

---

## Step 2: Command Bar

**Goal:** Add a `:` command bar at the bottom of the screen for view navigation.

### 2a. Add command bar to mainModel

**File:** `ui.go`

Add fields to `mainModel`:
```go
type mainModel struct {
    state          tabIndex
    width          int
    height         int
    tabs           []string
    tabContent     []tea.Model
    loaded         bool
    // NEW: command bar
    commandBar       textinput.Model
    commandBarActive bool
}
```

In `newModel()`, initialize:
```go
cb := textinput.New()
cb.Placeholder = "type a command..."
cb.CharLimit = 64
cb.SetSuggestions([]string{"releases", "repos", "hub", "plugins", "quit"})
cb.ShowSuggestions = true
m.commandBar = cb
```

### 2b. Handle `:` key activation

**File:** `ui.go` — in the `tea.KeyMsg` switch inside `Update()`:

```go
case ":":
    if !m.commandBarActive {
        m.commandBarActive = true
        m.commandBar.Focus()
        m.commandBar.SetValue("")
        return m, textinput.Blink
    }
```

When command bar is active, intercept ALL key messages before dispatching to tabs:
```go
if m.commandBarActive {
    switch msg.String() {
    case "enter":
        m.commandBarActive = false
        m.commandBar.Blur()
        cmd = m.executeCommand(m.commandBar.Value())
        return m, cmd
    case "esc":
        m.commandBarActive = false
        m.commandBar.Blur()
        return m, nil
    }
    m.commandBar, cmd = m.commandBar.Update(msg)
    return m, cmd
}
```

### 2c. Implement executeCommand()

**File:** `ui.go` — new function:

```go
var commandAliases = map[string]tabIndex{
    "releases": releasesTab,
    "rel":      releasesTab,
    "repos":    repositoriesTab,
    "repo":     repositoriesTab,
    "repositories": repositoriesTab,
    "hub":      hubTab,
    "plugins":  pluginsTab,
    "plug":     pluginsTab,
}

func (m *mainModel) executeCommand(input string) tea.Cmd {
    input = strings.TrimSpace(strings.ToLower(input))
    if input == "q" || input == "quit" {
        return tea.Quit
    }
    if target, ok := commandAliases[input]; ok {
        m.state = target
    }
    return nil
}
```

### 2d. Suppress `[`/`]` and `:` propagation to child tabs

Currently `[`/`]` switch tabs AND the key is forwarded to the active tab. After switching to command-bar navigation:
- Keep `[`/`]` as quick tab shortcuts (backward compat) but do NOT forward them to child tabs — return early after state change.
- `:` must NOT propagate to child tabs (it activates the command bar).

**File:** `ui.go` — in the KeyMsg handler, add `return m, nil` after `[`/`]` state changes, and add `:` case before tab dispatch.

### 2e. Adjust height for command bar

When the command bar is active (or always, to reserve space), subtract its height from the content area:

**File:** `ui.go` — in `View()` and in the `WindowSizeMsg` handler, compute:
```go
chromeHeight := lipgloss.Height(m.renderHeader()) + 1 // +1 for status/command bar
```

Pass `msg.Height - chromeHeight` to tab sub-models instead of `msg.Height - lipgloss.Height(m.renderMenu())`.

---

## Step 3: Breadcrumb Header

**Goal:** Replace the tab menu with a breadcrumb bar showing navigation path.

### 3a. Add breadcrumb state

**File:** `ui.go`

Add to `mainModel`:
```go
type mainModel struct {
    // ... existing fields ...
    breadcrumbs []string  // e.g., ["Releases"] or ["Releases", "nginx", "Values"]
}
```

### 3b. New message type for breadcrumb updates

**File:** `types/messages.go`

```go
type BreadcrumbMsg struct {
    Crumbs []string
}
```

Each tab module emits this when the user drills into a sub-view. For example, in `releases/overview.go` when entering detail view:
```go
case "enter", " ":
    // ... existing drill-down logic ...
    cmds = append(cmds, func() tea.Msg {
        return types.BreadcrumbMsg{Crumbs: []string{m.releaseTable.SelectedRow()[0]}}
    })
```

When pressing Esc to go back:
```go
case "esc":
    // ... existing back logic ...
    cmds = append(cmds, func() tea.Msg {
        return types.BreadcrumbMsg{Crumbs: []string{}}
    })
```

### 3c. Replace renderMenu() with renderHeader()

**File:** `ui.go`

Replace `renderMenu()` with:
```go
func (m mainModel) renderHeader() string {
    // Left side: breadcrumb path
    crumbs := []string{m.tabs[m.state]}
    crumbs = append(crumbs, m.breadcrumbs...)
    path := strings.Join(crumbs, " > ")
    pathStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.HighlightColor)
    left := pathStyle.Render(path)

    // Right side: tab indicators (dimmed, active one highlighted)
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

    // Fill the gap
    gap := max(0, m.width - lipgloss.Width(left) - lipgloss.Width(right))
    return left + strings.Repeat(" ", gap) + right
}
```

### 3d. Handle BreadcrumbMsg in Update()

**File:** `ui.go`:
```go
case types.BreadcrumbMsg:
    m.breadcrumbs = msg.Crumbs
```

### 3e. Emit breadcrumbs from each module

Each module emits `BreadcrumbMsg` when drilling into/out of sub-views:

**`releases/overview.go`:**
- On `enter`/`space` (drill into release detail): emit `BreadcrumbMsg{Crumbs: []string{releaseName, "History"}}`
- On `h`/`l` tab switch in detail view: emit `BreadcrumbMsg{Crumbs: []string{releaseName, viewName}}`
- On `esc` back to overview: emit `BreadcrumbMsg{Crumbs: []string{}}`

**`repositories/overview.go`:**
- On `l`/`right` panel switch: emit crumb for current panel focus
- On `esc`: emit empty crumbs

**`hub/hub.go`:**
- On entering default value view: emit `BreadcrumbMsg{Crumbs: []string{packageName, "Values"}}`
- On `esc`: emit empty crumbs

**`plugins/overview.go`:**
- No sub-views, no breadcrumb updates needed

---

## Step 4: Status Bar

**Goal:** Replace per-view help bars with a unified status bar at the bottom.

### 4a. New message type

**File:** `types/messages.go`

```go
type StatusMsg struct {
    Text string  // e.g., "Release nginx deleted", "42 releases"
}
```

### 4b. Add status bar state to mainModel

**File:** `ui.go`

```go
type mainModel struct {
    // ... existing fields ...
    statusText     string    // current status message
    statusFlash    string    // temporary flash message
    statusFlashCmd tea.Cmd   // timer to clear flash
}
```

### 4c. Add flash message type and clearing

**File:** `types/messages.go`

```go
type ClearFlashMsg struct{}
```

**File:** `ui.go` — handle in Update:
```go
case types.StatusMsg:
    m.statusFlash = msg.Text
    return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
        return types.ClearFlashMsg{}
    })
case types.ClearFlashMsg:
    m.statusFlash = ""
```

### 4d. Render status bar

**File:** `ui.go`

```go
func (m mainModel) renderStatusBar() string {
    left := ""
    if m.commandBarActive {
        left = ":" + m.commandBar.View()
    } else if m.statusFlash != "" {
        left = lipgloss.NewStyle().Foreground(lipgloss.Color("green")).Render(m.statusFlash)
    }

    right := m.statusText
    gap := max(0, m.width - lipgloss.Width(left) - lipgloss.Width(right))
    return left + strings.Repeat(" ", gap) + right
}
```

### 4e. Remove per-view help rendering

Each view currently appends `helpView` at the bottom. Move this responsibility to the main model.

**Files to modify:**
- `releases/overview_view.go:39` — remove `helpView` from `View()` return
- `repositories/overview_view.go:26` — remove `helpView` concatenation
- `hub/hub_view.go:24-39` — remove `helpView` concatenation
- `plugins/overview_view.go:16,22` — remove `helpView` concatenation

Each module's `View()` should return ONLY its content (tables, viewports, etc.) without the help bar. The main model's `View()` renders: `header + content + statusBar`.

### 4f. Show context-sensitive key hints

The status bar right side shows key hints. Each module exports a method `KeyHints() string` that returns abbreviated hints for the current state:

```go
// Example for releases in overview mode:
func (m Model) KeyHints() string {
    return "i:install  D:delete  u:upgrade  r:refresh  enter:details  ?:help"
}
```

The main model calls this on the active tab and displays it in the status bar right section. This replaces the bottom help.View() pattern.

**Alternative (simpler):** Keep `help.Model` but render it from the main model instead of each sub-model. The main model would need access to the active tab's keymap. Since tab models return `tea.Model` (interface), add a `KeyMap() key.Map` method to a new interface:

```go
type KeyHintProvider interface {
    KeyHints() string
}
```

Each tab model implements it. The main model type-asserts and calls it.

---

## Step 5: View Back-Stack

**Goal:** Unified history navigation with Escape = back.

### 5a. Define view state

**File:** `ui.go`

```go
type viewState struct {
    tab        tabIndex
    breadcrumbs []string
}
```

### 5b. Add history stack to mainModel

```go
type mainModel struct {
    // ... existing fields ...
    history     []viewState
    historyIdx  int  // current position in history (-1 = none)
}
```

### 5c. Push state on navigation

When the user switches tabs (via command bar or `[`/`]`), push the current state:
```go
func (m *mainModel) pushHistory() {
    state := viewState{tab: m.state, breadcrumbs: m.breadcrumbs}
    // Trim forward history if we navigated back then went elsewhere
    m.history = append(m.history[:m.historyIdx+1], state)
    m.historyIdx = len(m.history) - 1
    // Cap at 50
    if len(m.history) > 50 {
        m.history = m.history[len(m.history)-50:]
        m.historyIdx = len(m.history) - 1
    }
}
```

### 5d. Navigate history with `[`/`]`

Repurpose `[`/`]` from tab cycling to history navigation:
```go
case "]":
    if m.historyIdx < len(m.history)-1 {
        m.historyIdx++
        state := m.history[m.historyIdx]
        m.state = state.tab
        m.breadcrumbs = state.breadcrumbs
    }
case "[":
    if m.historyIdx > 0 {
        m.historyIdx--
        state := m.history[m.historyIdx]
        m.state = state.tab
        m.breadcrumbs = state.breadcrumbs
    }
```

### 5e. Keep Escape as within-view back

Escape is handled by each module (e.g., releases drill-down back to overview). This stays as-is — it's intra-view navigation. `[`/`]` is inter-view (cross-tab) navigation.

---

## Step 6: Update View() Assembly

**File:** `ui.go` — rewrite `View()`:

```go
func (m mainModel) View() string {
    if !m.loaded || len(m.tabContent) == 0 {
        return "loading..."
    }
    header := m.renderHeader()
    content := m.tabContent[m.state].View()
    statusBar := m.renderStatusBar()
    return header + "\n" + content + "\n" + statusBar
}
```

Update `WindowSizeMsg` to subtract chrome height (header + status bar = 2 lines):
```go
chromeHeight := 2 // header + status bar
adjustedMsg := tea.WindowSizeMsg{Width: m.width, Height: msg.Height - chromeHeight}
```

---

## Files to Modify

| File | Changes |
|------|---------|
| `ui.go` | Major rewrite: add commandBar, breadcrumbs, statusBar, history stack, replace renderMenu with renderHeader, new renderStatusBar, executeCommand, pushHistory, rewrite Update/View |
| `main.go` | No changes |
| `components/table.go` | Remove 6 Unbind() lines in GenerateTable(), add title param to RenderTable() |
| `types/messages.go` | Add BreadcrumbMsg, StatusMsg, ClearFlashMsg |
| `helpers/keymaps.go` | Update CommonKeys: remove `[`/`]` "Change panel" hint, add `:` "Command" hint, `?` "Help" hint |
| `styles/styles.go` | No changes (use existing styles) |
| `styles/helpers.go` | No changes |
| `releases/overview.go` | Emit BreadcrumbMsg on drill/back |
| `releases/overview_view.go` | Remove helpView from View() return |
| `releases/overview_keymap.go` | No changes (keep existing bindings) |
| `repositories/overview.go` | Emit BreadcrumbMsg on panel switch/back |
| `repositories/overview_view.go` | Remove helpView from View() return |
| `hub/hub.go` | Emit BreadcrumbMsg on view changes |
| `hub/hub_view.go` | Remove helpView from View() return |
| `plugins/overview.go` | No changes |
| `plugins/overview_view.go` | Remove helpView from View() return, update RenderTable call with title |

**Total: 14 files modified, 0 new files**

## Implementation Order

```
Step 0 (bug fixes) ─── no dependencies
Step 1 (vim keys)  ─── no dependencies
Step 2 (command bar) ── no dependencies
Step 3 (breadcrumbs) ── depends on Step 2 (header layout)
Step 4 (status bar) ─── depends on Step 3 (chrome height calc)
Step 5 (back-stack) ─── depends on Steps 2-4
Step 6 (assembly) ──── depends on all above
```

Parallelizable: Steps 0, 1, and 2 can all be done independently.

## Acceptance Criteria

- [ ] `go build` succeeds with zero errors
- [ ] `go vet ./...` passes
- [ ] `go test ./...` passes (existing tests)
- [ ] App launches and shows header with breadcrumbs + tab indicators
- [ ] Pressing `:` opens command bar at bottom
- [ ] Typing `releases` + Enter in command bar switches to Releases tab
- [ ] Typing `repos` + Enter switches to Repositories tab
- [ ] Tab autocomplete works in command bar
- [ ] `:q` quits the app
- [ ] `Escape` dismisses command bar without action
- [ ] `j`/`k` move up/down in all table views
- [ ] `g`/`G` jump to top/bottom in all tables
- [ ] `Ctrl-f`/`Ctrl-b` page down/up in all tables
- [ ] Breadcrumbs update when drilling into release details
- [ ] Breadcrumbs reset when pressing Escape back to overview
- [ ] Status bar shows at bottom of screen
- [ ] `[`/`]` navigate view history (back/forward)
- [ ] All existing functionality works: install wizard, upgrade wizard, delete, rollback, repo management, hub search, plugin management
- [ ] No regressions in window resize behavior
