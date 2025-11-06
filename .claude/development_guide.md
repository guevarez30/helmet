# Dockit - Development Guide

## For Future Claude Instances

This guide helps you understand how to work on Dockit effectively.

---

## Getting Started

### Build & Run
```bash
# Build
go build -buildvcs=false -o dockit

# Run
./dockit

# Clean build
go clean && go build -buildvcs=false -o dockit
```

### Common Issues
- **Missing dependencies**: Run `go mod tidy`
- **Build errors about types**: Check Docker SDK import paths
- **VCS errors**: Use `-buildvcs=false` flag

---

## Code Style & Patterns

### Bubble Tea Model Pattern

Every view follows this structure:

```go
// 1. Model struct
type MyViewModel struct {
    client *docker.Client
    data   []SomeType
    cursor int
    err    error
    keys   KeyMap
}

// 2. Constructor
func NewMyViewModel(client *docker.Client) *MyViewModel {
    return &MyViewModel{
        client: client,
        keys:   DefaultKeyMap(),
    }
}

// 3. Init - Load initial data
func (m *MyViewModel) Init() tea.Cmd {
    return m.refresh()
}

// 4. Update - Handle messages
func (m *MyViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // Handle keys
    case myDataMsg:
        // Update state
    }
    return m, nil
}

// 5. View - Render UI
func (m *MyViewModel) View() string {
    // Build UI string
    return lipgloss.JoinVertical(...)
}

// 6. Helper - Fetch data
func (m *MyViewModel) refresh() tea.Cmd {
    return func() tea.Msg {
        data, err := m.client.SomeMethod()
        if err != nil {
            return errMsg(err)
        }
        return myDataMsg(data)
    }
}
```

### Message Types

Always define typed messages:

```go
// Data messages
type containersMsg []types.Container

// Action messages
type containerActionMsg struct {
    success bool
    message string
}

// UI messages
type clearStatusMsg struct{}
```

---

## Adding New Features

### Adding a New View

1. **Create file** `ui/myview.go`

2. **Define model**:
```go
type MyViewModel struct {
    client *docker.Client
    // ... state fields
}
```

3. **Implement Init/Update/View**

4. **Add to main model** in `ui/model.go`:
```go
type View int
const (
    // ...
    MyNewView View = iota
)

type Model struct {
    // ...
    myView *MyViewModel
}
```

5. **Add routing** in `model.go Update()`:
```go
case MyNewView:
    newView, cmd := m.myView.Update(msg)
    m.myView = newView.(*MyViewModel)
    return m, cmd
```

6. **Add rendering** in `model.go View()`:
```go
case MyNewView:
    content = m.myView.View()
```

7. **Add to tabs** in `renderTabs()`:
```go
{"My View", MyNewView},
```

### Adding Docker Operations

1. **Add method to** `docker/client.go`:
```go
func (c *Client) MyOperation(id string) error {
    return c.cli.SomeDockerMethod(c.ctx, id, options)
}
```

2. **Use in view model**:
```go
func (m *MyViewModel) doAction() tea.Cmd {
    return func() tea.Msg {
        err := m.client.MyOperation(id)
        if err != nil {
            return errMsg(err)
        }
        return actionMsg{success: true}
    }
}
```

### Adding Keybindings

1. **Define in** `ui/keys.go`:
```go
type KeyMap struct {
    // ...
    MyAction key.Binding
}

func DefaultKeyMap() KeyMap {
    return KeyMap{
        // ...
        MyAction: key.NewBinding(
            key.WithKeys("a"),
            key.WithHelp("a", "my action"),
        ),
    }
}
```

2. **Handle in view Update()**:
```go
case key.Matches(msg, m.keys.MyAction):
    return m, m.doMyAction()
```

3. **Document in footer**:
```go
func (m Model) renderFooter() string {
    helpText := "... • a: my action"
    // ...
}
```

---

## Styling Guidelines

### Using Lipgloss

```go
// Define styles in ui/styles.go
var MyStyle = lipgloss.NewStyle().
    Foreground(primaryColor).
    Bold(true).
    Padding(0, 2)

// Use in views
content := MyStyle.Render("Hello")
```

### Color Usage
- **Purple** (`primaryColor`) - Active elements, highlights
- **Green** (`successColor`) - Running, success
- **Red** (`errorColor`) - Stopped, errors
- **Yellow** (`warningColor`) - Warnings, highlights
- **Cyan** (`infoColor`) - Labels, info text
- **Gray** (`mutedColor`) - Help text, inactive

### Layout Patterns

**Tabular Data**:
```go
row := fmt.Sprintf("%-20s  %-30s  %s", col1, col2, col3)
```

**Vertical Stacking**:
```go
lipgloss.JoinVertical(lipgloss.Left, line1, line2, line3)
```

**Horizontal Layout**:
```go
lipgloss.JoinHorizontal(lipgloss.Top, left, right)
```

---

## Common Tasks

### Adding Visual Feedback

```go
type MyModel struct {
    statusMsg        string
    actionInProgress bool
}

func (m *MyModel) doAction() tea.Cmd {
    m.actionInProgress = true
    return func() tea.Msg {
        // Do work
        return actionMsg{message: "Success"}
    }
}

func (m *MyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    case actionMsg:
        m.statusMsg = msg.message
        m.actionInProgress = false
        return m, tea.Batch(
            m.refresh(),
            m.clearStatusAfter(2 * time.Second),
        )
}

func (m *MyModel) View() string {
    if m.actionInProgress {
        // Show "⟳ Processing..."
    }
    if m.statusMsg != "" {
        // Show "✓ Success message"
    }
    // ... rest of view
}
```

### Handling Lists

```go
case key.Matches(msg, m.keys.Up):
    if m.cursor > 0 {
        m.cursor--
    }
case key.Matches(msg, m.keys.Down):
    if m.cursor < len(m.items)-1 {
        m.cursor++
    }
```

### Rendering Selected Items

```go
for i, item := range m.items {
    row := renderRow(item, i == m.cursor)
    rows = append(rows, row)
}

func renderRow(item Item, selected bool) string {
    row := formatItem(item)
    if selected {
        return lipgloss.NewStyle().
            Background(primaryColor).
            Foreground(lipgloss.Color("#FAFAFA")).
            Render(row)
    }
    return row
}
```

---

## Testing Approach

### Manual Testing Checklist

**Containers View**:
- [ ] List shows all containers
- [ ] Start works on stopped container
- [ ] Stop works on running container
- [ ] Restart works and uptime resets
- [ ] Remove works and container disappears
- [ ] Status messages appear and clear
- [ ] Navigation (up/down) works
- [ ] Tab switching works

**Images View**:
- [ ] List shows all images
- [ ] Size displays correctly
- [ ] Dangling images marked
- [ ] Remove works
- [ ] Navigation works

**Logs View**:
- [ ] Logs load and display
- [ ] Scrolling works
- [ ] Search (/) opens input
- [ ] Search filters correctly
- [ ] Matches highlighted
- [ ] Esc clears filter
- [ ] Esc exits logs

### Docker Test Setup

```bash
# Create test containers
docker run -d --name test-nginx nginx
docker run -d --name test-redis redis
docker pull alpine
```

---

## Debugging Tips

### Bubble Tea Debugging

**Don't use fmt.Println** - it breaks TUI

Instead:
1. Log to file:
```go
f, _ := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
fmt.Fprintf(f, "Debug: %v\n", value)
```

2. Show in UI:
```go
m.debugMsg = fmt.Sprintf("Debug: %v", value)
// Render m.debugMsg in View()
```

### Common Bugs

**Tabs not showing**:
- Check if tabs are rendered in main model View()
- Verify tab rendering not skipped for current view

**Logs stuck loading**:
- Ensure `m.ready = true` set when data arrives
- Check logsMsg being received in Update()

**Text wrapping**:
- Use fixed-width columns: `fmt.Sprintf("%-20s", text)`
- Truncate: `if len(text) > 20 { text = text[:17] + "..." }`

**Keys not working**:
- Check key.Matches() uses correct key binding
- Verify not in different mode (e.g., search mode)

---

## Docker Client Gotchas

### Type Changes
Docker SDK types can change between versions:
- Old: `types.ContainersPruneConfig`
- New: `filters.Args`

**Solution**: Check SDK docs or examples

### API Contexts
Always use `c.ctx` for cancellation:
```go
c.cli.ContainerList(c.ctx, options)
```

### Log Format
Docker logs have 8-byte headers. Always parse or strip them.

---

## Performance Notes

### Current Limitations
- Blocking operations freeze UI
- Full refresh on every action
- No background polling

### Future Improvements
- Use tea.Tick for periodic updates
- Batch operations
- Virtual scrolling for huge lists
- Streaming logs

---

## File Organization

```
ui/
├── model.go        # Main app model, routing, tabs
├── dashboard.go    # Stats cards
├── containers.go   # Container list, operations
├── images.go       # Image list, operations
├── logs.go         # Log viewer, search
├── styles.go       # Lipgloss styles, colors
└── keys.go         # Key bindings, help
```

**Rule of thumb**:
- One file per view
- Styles centralized in styles.go
- Keys centralized in keys.go
- Main model coordinates everything

---

## Pull Request Guidelines (Future)

### Commit Messages
```
feat: add volume management view
fix: correct log parsing for multiline output
refactor: extract common list rendering
docs: update keybinding reference
```

### Before Submitting
- [ ] Code builds without warnings
- [ ] Tested manually with real containers
- [ ] No debug code left in
- [ ] Updated README if user-facing
- [ ] Added to .claude docs if architectural

---

## Helpful Resources

- **Bubble Tea docs**: https://github.com/charmbracelet/bubbletea
- **Lipgloss examples**: https://github.com/charmbracelet/lipgloss/tree/master/examples
- **Docker SDK docs**: https://pkg.go.dev/github.com/docker/docker/client
- **Bubble Tea examples**: https://github.com/charmbracelet/bubbletea/tree/master/examples

---

## Quick Reference

### Rebuild & Test
```bash
go build -buildvcs=false -o dockit && ./dockit
```

### Check Types
```bash
go doc github.com/docker/docker/api/types
```

### Format Code
```bash
gofmt -w .
```

### Dependencies
```bash
go mod tidy
go mod download
```
