# K9s UX Model Reference

## Overview

K9s is a terminal-based UI for Kubernetes cluster management. It uses a vim-inspired, command-driven navigation model with colon command bar, slash filter bar, breadcrumb navigation, and rich keyboard shortcuts.

## 1. Navigation Model

### Core Paradigm

- Command-driven: Users navigate via `:` command bar
- Hierarchical: cluster > context > namespace > resource type > resource > sub-views
- Vim-inspired: j/k/h/l movement, modal input

### How Users Move

| Mechanism | How |
|---|---|
| `:` command bar | Type resource name (pod, deploy, svc), press Enter to jump |
| Tab / Ctrl-f | Accept autocomplete suggestion in command bar |
| Enter | Drill into resource (pod -> containers) |
| Escape | Back to previous view |
| Breadcrumbs | Header shows cluster/namespace/resource path (toggle: Ctrl+e) |
| `[` / `]` | Navigate command history backward/forward |
| `-` | Return to last active command |
| `0-9` | Quick namespace favorites |
| Hotkeys | User-defined shortcuts (hotkeys.yaml) |

### Context & Namespace Switching

- `:ctx` lists contexts, `:ctx name` switches
- `:pod @staging` views pods in alternate context
- `:ns` lists namespaces, number keys 0-9 for favorites
- `0` shows all namespaces, `u` marks favorite
- `:pod kube-system` views pods in specific namespace

## 2. Search & Filter System

Four filter modes, all triggered with `/`:

| Mode | Syntax | Description |
|---|---|---|
| Regex | `/pattern` | Standard regex matching (Regex2) |
| Inverse | `/!pattern` | Excludes matching resources |
| Label | `/-l label=value` | Kubernetes label selectors |
| Fuzzy | `/-f term` | Approximate/fuzzy matching |

- Filters are live (update as you type)
- Can combine with colon: `:pod /fred`
- Label selectors: `:pod app=fred,env=dev`
- Escape clears filter

## 3. Command Bar (`:`)

- Opens at bottom of screen
- Resource navigation: `:pod`, `:deploy`, `:svc`, `:ns`, `:node`, `:pvc`, `:secret`, `:cm`, `:crd`, `:events`, `:helm`
- With namespace: `:pod kube-system`
- With filter: `:pod /nginx`
- With labels: `:pod app=fred,env=dev`
- With context: `:pod @production`
- Special views: `:pulse`, `:xray deploy`, `:popeye`, `:screendump`, `:dir /path`
- Exit: `:q` or `:quit`
- Autocomplete with Tab/Ctrl-f
- Custom aliases supported (aliases.yaml)

## 4. Key Bindings

### Global

| Key | Action |
|---|---|
| `?` | Help (context-sensitive) |
| `:` | Command mode |
| `/` | Filter mode |
| `Ctrl+a` | Show all resource aliases |
| `Esc` | Back/exit mode |
| `:q` / `Ctrl+c` | Quit |
| `[` / `]` | Command history back/forward |
| `-` | Last active command |

### Vim Navigation

| Key | Action |
|---|---|
| `j` / `k` | Down / Up |
| `h` / `l` | Left / Right |
| `Ctrl+f` / `Ctrl+b` | Page down / up |
| `g` / `G` | Top / Bottom |
| `Enter` | Select / drill in |

### Resource Actions

| Key | Action |
|---|---|
| `d` | Describe resource |
| `v` | View resource (formatted) |
| `y` | View YAML |
| `e` | Edit in $EDITOR |
| `l` | View logs |
| `s` | Shell / Scale (context-dependent) |
| `a` | Apply resource |
| `r` | Restart / Auto-refresh |
| `x` | Decode secret |

### Destructive Actions

| Key | Action |
|---|---|
| `Ctrl+d` | Delete (with confirmation) |
| `Ctrl+k` | Force kill (no grace period) |
| `Ctrl+l` | Rollback deployment |

### Multi-Select

| Key | Action |
|---|---|
| `Space` | Toggle select |
| `Ctrl+a` | Select all visible |

### Display Controls

| Key | Action |
|---|---|
| `Ctrl+w` | Toggle wide columns |
| `Ctrl+e` | Toggle header |
| `f` | Fullscreen |
| `Ctrl+s` | Save to file (screendump) |
| `Ctrl+z` | Toggle error display |
| `Ctrl+r` | Manual refresh |

### Log View

| Key | Action |
|---|---|
| `w` | Toggle line wrap |
| `f` | Follow logs |
| `0` | All logs |
| `1-9` | Last N*100 lines |

### Sorting (Shift+letter)

| Key | Sorts By |
|---|---|
| `Shift+N` | Name |
| `Shift+P` | Namespace |
| `Shift+S` | Status |
| `Shift+A` | Age |
| `Shift+C` | CPU |
| `Shift+M` | Memory |
| `Shift+O` | Node |
| `Shift+I` | IP |
| `Shift+T` | Restarts |
| `Shift+R` | Readiness |

### Port Forwarding

| Key | Action |
|---|---|
| `Shift+F` | Port forward (dialog) |
| `Ctrl+B` | HTTP benchmark |

## 5. Screen Layout

```
+------------------------------------------------------------------+
| [Logo]  Context: prod  Cluster: my-cluster   [Key Hints]        |  <- Header
|  Pods < default >                                                |  <- Breadcrumbs
+------------------------------------------------------------------+
| NAMESPACE | NAME          | READY | STATUS  | RESTARTS | AGE    |  <- Column Headers
|-----------|---------------|-------|---------|----------|--------|
| default   | nginx-abc123  | 1/1   | Running | 0        | 2d     |  <- Table Body
+------------------------------------------------------------------+
| :                                                                |  <- Command/Status Bar
+------------------------------------------------------------------+
```

### Layout Components

- **Header**: Logo (togglable), cluster info, context-sensitive key hints
- **Breadcrumbs**: ResourceType < namespace > (togglable)
- **Table body**: Sortable, filterable, color-coded status
- **Command/status bar**: Bottom, shows : command or / filter input

### Status Colors

- Green: Healthy/Running
- Yellow: Warning/Pending
- Red: Error/Failed
- `+` marker: Favorite namespace
- `*` marker: Default namespace

### UI Config Flags

| Flag | Effect |
|---|---|
| headless | No splash screen |
| logoless | No K9s logo |
| crumbsless | No breadcrumbs |
| splashless | No splash on startup |
| noIcons | No terminal icons |
| enableMouse | Mouse support |
| invert | Dark -> light mode |

## 6. View Types

1. **Table View** (default): Sortable/filterable resource table, wide mode (Ctrl+w)
2. **Describe View** (`d`): Full kubectl describe output
3. **YAML View** (`y`): Raw YAML manifest with syntax highlighting
4. **Edit View** (`e`): Opens in $EDITOR, changes applied on save
5. **Log View** (`l`): Interactive log viewer (follow, wrap, tail length, timestamps)
6. **Shell View** (`s`): Interactive container shell
7. **XRay View** (`:xray RESOURCE`): Tree view of resource dependencies (Deployment->RS->Pod->Container->ConfigMaps)
8. **Pulse View** (`:pulse`): Real-time cluster health dashboard (CPU, memory, network)
9. **Popeye View** (`:popeye`): Cluster health analysis with ratings
10. **Screendump View** (`:screendump`): Previously saved outputs
11. **Dir View** (`:dir /path`): Local filesystem browser

## 7. Plugin System

### Architecture

Plugins map keyboard shortcuts to external shell commands with injected context variables.

### Plugin Definition

```yaml
plugins:
  plugin-name:
    shortCut: Ctrl-L
    description: "Tail pod logs"
    scopes: [po]
    command: kubectl
    background: false
    confirm: false
    args: [logs, -f, $NAME, -n, $NAMESPACE, --context, $CONTEXT]
```

### Available Environment Variables

$NAME, $NAMESPACE, $CONTEXT, $CLUSTER, $USER, $GROUPS, $KUBECONFIG, $CONTAINER, $POD, $FILTER, $RESOURCE_GROUP, $RESOURCE_VERSION, $RESOURCE_NAME, $COL-<COLUMN>

### Plugin Locations

1. $XDG_CONFIG_HOME/k9s/plugins.yaml
2. $XDG_CONFIG_HOME/k9s/plugins/ (directory of snippets)
3. $XDG_DATA_HOME/k9s/plugins/
4. Per-context: clusters/clusterX/contextY/plugins.yaml

## 8. Skin/Theme System

### Applying Skins

- Global: `k9s.ui.skin: theme_name` in config.yaml
- Per-context: skin attribute in context config
- Environment: `K9S_SKIN="dracula"`
- Default theme: Dracula

### Skin Structure

```yaml
k9s:
  body: { fgColor, bgColor, logoColor }
  info: { fgColor, sectionColor }
  frame:
    border: { fgColor, focusColor }
    menu: { fgColor, keyColor, numKeyColor }
    crumbs: { fgColor, bgColor, activeColor }
    status: { newColor, modifyColor, addColor, errorColor, highlightColor, killColor, completedColor }
    title: { fgColor, bgColor, highlightColor, counterColor, filterColor }
  views:
    table: { fgColor, bgColor, cursorColor, header: { fgColor, bgColor, sorterColor } }
    yaml: { keyColor, colonColor, valueColor }
    logs: { fgColor, bgColor }
  dialog: { ... }
```

### Color Support

- 140+ named colors, hex (#RRGGBB), `default` for transparent
- Common pattern: red skin = production, yellow = staging, green = dev

## 9. Configuration Files

| File | Purpose |
|---|---|
| config.yaml | Main config (UI, logger, shell, refresh rate) |
| aliases.yaml | Custom command shortcuts |
| hotkeys.yaml | Custom keyboard shortcuts |
| plugins.yaml | External command integration |
| views.yaml | Custom column layouts |
| skins/*.yaml | Theme files |
