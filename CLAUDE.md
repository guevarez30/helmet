# Helmet (helm-tui)

Terminal UI for managing Helm releases, charts, repositories, and plugins.
Built with Go 1.22 + BubbleTea (Charm ecosystem).

## Quick Reference

- **Build**: `make build` (output: `./bin/helm-tui`)
- **Run**: `make run`
- **Test**: `make test` or `go test ./...`
- **Format**: `make fmt`
- **Requires**: Go 1.22+, Helm 3 in PATH

## Architecture

MVU pattern (Model-View-Update) via BubbleTea. Root model in `ui.go` dispatches to 4 tab sub-models.

### Module Pattern

Each module follows a 4-file pattern:
- `overview.go` / `hub.go` -- Model struct, Init(), Update()
- `*_commands.go` -- tea.Cmd functions (shell to helm CLI)
- `*_keymap.go` -- Key bindings
- `*_view.go` -- View() render function

### Key Directories

| Directory | Purpose |
|-----------|---------|
| `releases/` | Helm releases (overview, install wizard, upgrade wizard) |
| `repositories/` | Helm repos (3-panel browser, add, install) |
| `hub/` | ArtifactHub search (HTTP API, not helm CLI) |
| `plugins/` | Helm plugin management |
| `components/` | Reusable table component with flex columns |
| `styles/` | Lipgloss styling, borders, colors |
| `helpers/` | Keymaps, constants, editor, logging |
| `types/` | Data structs (helm.go) and BubbleTea messages (messages.go) |

## Active Project: K9s UX Transformation

We are overhauling helmet to match k9s look, feel, and interaction patterns.

### Research & Planning

| Document | Path |
|----------|------|
| Helmet architecture map | `.claude/research/helmet-architecture.md` |
| K9s UX model reference | `.claude/research/k9s-ux-model.md` |
| Gap analysis & plan | `.claude/research/gap-analysis.md` |
| Learnings & gotchas | `.claude/notepads/learnings.md` |
| Design decisions | `.claude/notepads/decisions.md` |
| Issues & blockers | `.claude/notepads/issues.md` |

### Key Design Decisions

1. **Command bar (`:`)** replaces tab switching as primary navigation
2. **Filter bar (`/`)** added to all table views (regex, inverse, fuzzy)
3. **Vim keys** (j/k/g/G) for all navigation
4. **`[`/`]`** repurposed from tab cycling to command history
5. **UltraSearch** -- novel cross-view fuzzy search (beyond k9s)
6. **Skin system** -- YAML-based theming on top of lipgloss

### Known Bugs to Fix

1. `ui.go` WindowSizeMsg only batches plugins tab cmd (other 3 lost)
2. `components/table.go` RenderTable() hardcodes " Releases " title
3. `components/table.go` HalfPageDown unbound twice
4. `ui.go` mainModel.index field unused
