# Design Decisions

## Navigation Model
- **Decision**: Adopt k9s colon-command model as PRIMARY navigation, keep tab bar as SECONDARY visual indicator
- **Rationale**: The command bar is k9s's most distinctive UX feature. Users should be able to type `:releases` to jump views. The tab bar remains as a visual breadcrumb but `[`/`]` should become history navigation (matching k9s) not tab cycling.

## Filter Architecture
- **Decision**: Implement filter as a shared component injected into all table views
- **Rationale**: Each module currently has its own Update/View. A shared filter component that wraps any table.Model avoids duplicating filter logic in 4+ modules. The filter component handles `/` activation, mode detection (`/!`, `/-f`), and passes filtered rows to the table.

## Key Binding Strategy
- **Decision**: Remap to k9s conventions where possible
- **Rationale**: Users familiar with k9s should feel at home. Key changes:
  - `[`/`]` → command history (was: tab switch)
  - `:` → command bar (new)
  - `/` → filter (was: only in hub)
  - `j`/`k` → table up/down (was: arrows only)
  - `d` → describe view (new)
  - `y` → YAML view (was: accessed via tab navigation)
  - `?` → help panel (was: always visible at bottom)
  - `Ctrl+d` → delete with confirmation (was: `D`)
  - `Space` → multi-select (new)

## Styling Approach
- **Decision**: Keep Charm/lipgloss styling but add a skin system
- **Rationale**: lipgloss is already the rendering engine. A skin system loads color values from YAML and applies them to existing lipgloss styles. This is additive, not a rewrite.

## Message Architecture
- **Decision**: Keep existing message types, add new ones for new features
- **Rationale**: The current types/messages.go pattern works. New features (filter, command bar, sort) need new message types but the architecture doesn't need to change.

## UltraSearch Concept
- **Decision**: UltraSearch will be a cross-view fuzzy search overlay
- **Rationale**: This is the "beyond k9s" feature. When activated (e.g., `Ctrl+/` or `:search`), it opens a fullscreen overlay that searches across releases, repos, hub, and plugins simultaneously. Results are grouped by type with jump-to navigation. This does NOT exist in k9s and would be a differentiator.
