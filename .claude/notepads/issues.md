# Issues & Blockers

## Must Fix Before K9s Overhaul
1. **WindowSizeMsg bug** — Only plugins tab gets resize command batched. ALL tabs need resize. Fix: accumulate cmds in a slice, batch at end.
2. **Hardcoded table title** — RenderTable() says " Releases " for all tabs. Fix: Accept title parameter.
3. **Unused index field** — Remove mainModel.index to reduce confusion.

## Architecture Risks
1. **Global mutable state** — helpers.UserDir, helpers.LogFile, styles.WindowSize are global vars. Not blocking but makes testing harder. Consider passing via model fields.
2. **No config system** — No config file support means skin/theme/plugin features need a config loading system built from scratch.
3. **Message coupling** — Messages carry []table.Row instead of domain types. This works but makes the filter system harder (filtering needs access to raw field values, not pre-formatted row strings).
4. **Hardcoded helm path** — Commands assume `helm` is in PATH. Should support configurable helm binary path.

## Open Questions
1. Should command bar completely replace tab bar, or coexist?
   - **Current answer**: Coexist — command bar is primary, tab bar is visual indicator
2. How to handle the Hub module's HTTP API differently from other modules' CLI approach?
   - **Current answer**: Keep as-is, Hub is naturally different
3. Should filter work on the formatted table rows or the underlying data?
   - **Current answer**: Filter on formatted rows (simpler, matches k9s behavior)
4. What is the right UltraSearch activation key?
   - **Current answer**: TBD — candidates: Ctrl+/, Ctrl+Space, or `:search` command
