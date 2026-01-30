# Learnings

## Architecture Patterns
- Helmet uses BubbleTea MVU (Model-View-Update) pattern
- Each module follows 4-file pattern: model, commands, keymap, view
- All Helm operations shell out to `helm` CLI via exec.Command
- Hub is the only module using HTTP API (ArtifactHub)
- Messages defined centrally in types/messages.go
- Table component uses flex-column layout (components/table.go)

## Known Bugs Found
1. ui.go:100-120 — WindowSizeMsg handler only batches plugins tab resize cmd (overwrites other 3)
2. components/table.go:80 — RenderTable() hardcodes " Releases " title for ALL tabs
3. components/table.go:63 — HalfPageDown unbound twice (duplicate line)
4. ui.go:30 — mainModel.index field declared but never used
5. types/helm.go — History.Revision is int, Release.Revision is string (inconsistent)

## Gotchas
- Tab navigation uses `[` and `]` — these are also common vim keys, will conflict with k9s-style nav
- The Hub module has a different search input pattern than other modules
- Editor integration uses tea.ExecProcess which takes over the terminal
- Debounce pattern in hub/install uses Tag int field on DebounceEndMsg
- No configuration file system exists — everything is hardcoded
- helpers.UserDir and helpers.LogFile are global mutable state (not ideal but functional)

## BubbleTea Patterns
- tea.Cmd functions return tea.Msg when async work completes
- tea.Batch() combines multiple Cmds
- tea.ExecProcess() for running external editors (blocks TUI)
- textinput.Model for text fields, table.Model for tables, viewport.Model for scrollable content
- help.Model for rendering keybinding help
- key.Binding for defining keyboard shortcuts with help text
