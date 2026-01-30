package plugins

import (
	"github.com/pidanou/helm-tui/components"
	"github.com/pidanou/helm-tui/styles"
)

func (m PluginsModel) View() string {
	var remainingHeight = m.height
	if m.installPluginInput.Focused() {
		remainingHeight -= 3
	}
	view := components.RenderTable(m.pluginsTable, remainingHeight-3, m.width-2, " Plugins ")
	m.installPluginInput.Width = m.width - 5
	if m.installPluginInput.Focused() {
		view += "\n" + styles.ActiveStyle.Border(styles.Border).Render(m.installPluginInput.View())
	}
	return view
}
