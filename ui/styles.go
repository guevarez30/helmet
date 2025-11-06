package ui

import "github.com/charmbracelet/lipgloss"

// Color palette - Kubernetes inspired
const (
	k8sBlue       = "#326CE5" // Kubernetes primary blue
	primaryColor  = "#7D56F4" // Purple - active elements
	successColor  = "#50FA7B" // Green - deployed, success
	warningColor  = "#FFB86C" // Orange - pending, warnings
	errorColor    = "#FF5555" // Red - failed, errors
	infoColor     = "#8BE9FD" // Cyan - info, labels
	mutedColor    = "#6272A4" // Gray - inactive, help
	bgColor       = "#282A36" // Background
)

var (
	// Base styles
	BaseStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	// Title styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(k8sBlue)).
			MarginBottom(1)

	// Tab styles
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color(k8sBlue)).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(mutedColor)).
				Padding(0, 2)

	TabSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(mutedColor))

	// Status bar styles
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(infoColor)).
			Background(lipgloss.Color(bgColor)).
			Padding(0, 1)

	ContextStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(k8sBlue)).
			Bold(true)

	NamespaceStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(primaryColor)).
			Bold(true)

	// Card styles (for dashboard)
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(k8sBlue)).
			Padding(1, 2).
			MarginRight(2).
			MarginBottom(1)

	CardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(k8sBlue)).
			MarginBottom(1)

	// Status styles
	DeployedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(successColor)).
			Bold(true)

	FailedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(errorColor)).
			Bold(true)

	PendingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(warningColor)).
			Bold(true)

	UninstallingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(warningColor))

	UnknownStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(mutedColor))

	// List styles
	SelectedItemStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(k8sBlue)).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true)

	UnselectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(infoColor)).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(lipgloss.Color(mutedColor))

	// Message styles
	SuccessMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(successColor)).
				Bold(true)

	ErrorMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(errorColor)).
				Bold(true)

	InfoMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(infoColor))

	ProcessingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(warningColor))

	// Help text style
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(mutedColor)).
			MarginTop(1)

	// Search/input styles
	InputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(k8sBlue)).
			Padding(0, 1)

	InputFocusedStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(primaryColor)).
				Padding(0, 1)

	// Chart/release version style
	VersionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(infoColor))

	// Highlight style (for search results)
	HighlightStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(warningColor)).
			Foreground(lipgloss.Color("#000000"))

	// Separator style
	SeparatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(mutedColor)).
			MarginTop(1).
			MarginBottom(1)

	// Dialog styles
	DialogStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))
)

// GetStatusStyle returns the appropriate style for a release status
func GetStatusStyle(status string) lipgloss.Style {
	switch status {
	case "deployed":
		return DeployedStyle
	case "failed":
		return FailedStyle
	case "pending", "pending-install", "pending-upgrade", "pending-rollback":
		return PendingStyle
	case "uninstalling":
		return UninstallingStyle
	default:
		return UnknownStyle
	}
}

// StatusIndicator returns a colored circle indicator for status
func StatusIndicator(status string) string {
	switch status {
	case "deployed":
		return DeployedStyle.Render("●")
	case "failed":
		return FailedStyle.Render("●")
	case "pending", "pending-install", "pending-upgrade", "pending-rollback":
		return PendingStyle.Render("●")
	default:
		return UnknownStyle.Render("●")
	}
}

// PodStatusIndicator returns a colored circle indicator for pod status
func PodStatusIndicator(status string) string {
	switch status {
	case "Running":
		return DeployedStyle.Render("●")
	case "Succeeded":
		return DeployedStyle.Render("●")
	case "Failed":
		return FailedStyle.Render("●")
	case "Error", "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull":
		return FailedStyle.Render("●")
	case "Pending", "ContainerCreating", "PodInitializing":
		return PendingStyle.Render("●")
	case "NotReady", "Terminating":
		return PendingStyle.Render("●")
	case "Unknown":
		return UnknownStyle.Render("●")
	default:
		return UnknownStyle.Render("●")
	}
}

// GetPodStatusStyle returns the appropriate style for a pod status
func GetPodStatusStyle(status string) lipgloss.Style {
	switch status {
	case "Running", "Succeeded":
		return DeployedStyle
	case "Failed", "Error", "CrashLoopBackOff", "ImagePullBackOff", "ErrImagePull":
		return FailedStyle
	case "Pending", "ContainerCreating", "PodInitializing", "NotReady", "Terminating":
		return PendingStyle
	default:
		return UnknownStyle
	}
}

// Truncate truncates a string to a maximum length with ellipsis
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max < 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

// PadRight pads a string to a fixed width
func PadRight(s string, width int) string {
	if len(s) >= width {
		return Truncate(s, width)
	}
	return s + lipgloss.NewStyle().Width(width - len(s)).Render("")
}
