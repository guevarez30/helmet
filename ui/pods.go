package ui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	corev1 "k8s.io/api/core/v1"
)

// PodsModel represents the pods view for a release
type PodsModel struct {
	releaseName string
	pods        []corev1.Pod
	cursor      int
	loading     bool
	err         error
}

// NewPodsModel creates a new pods model
func NewPodsModel() *PodsModel {
	return &PodsModel{
		loading: false,
	}
}

// Init initializes the pods view
func (m *PodsModel) Init() tea.Cmd {
	return nil
}

// Update handles pods view messages
func (m *PodsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case podsMsg:
		m.pods = msg.pods
		m.releaseName = msg.releaseName
		m.loading = false
		return m, nil

	case errMsg:
		m.err = error(msg)
		m.loading = false
		return m, nil
	}

	return m, nil
}

// View renders the pods view
func (m *PodsModel) View() string {
	if m.loading {
		return ProcessingStyle.Render("⟳ Loading pods...")
	}

	if m.err != nil {
		return ErrorMessageStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var b strings.Builder

	// Title
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Pods for Release: %s (%d)", m.releaseName, len(m.pods))))
	b.WriteString("\n\n")

	// Pods table
	if len(m.pods) == 0 {
		b.WriteString(InfoMessageStyle.Render("No pods found for this release"))
		b.WriteString("\n\n")
	} else {
		b.WriteString(m.renderPodsTable())
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓: navigate • l: logs • esc: back • q: quit"))

	return b.String()
}

// renderPodsTable renders the pods table
func (m *PodsModel) renderPodsTable() string {
	var b strings.Builder

	// Header
	header := fmt.Sprintf("%-40s %-20s %-10s %-15s %-20s",
		"NAME", "STATUS", "RESTARTS", "READY", "AGE")
	b.WriteString(HeaderStyle.Render(header))
	b.WriteString("\n")

	// Rows
	for i, pod := range m.pods {
		name := Truncate(pod.Name, 40)
		status := m.getPodStatus(&pod)
		restarts := m.getPodRestarts(&pod)
		ready := m.getPodReady(&pod)
		age := m.formatAge(pod.CreationTimestamp.Time)

		// Format status with indicator and color
		statusIndicator := PodStatusIndicator(status)
		statusText := Truncate(status, 17)
		formattedStatus := fmt.Sprintf("%s %s", statusIndicator, statusText)

		row := fmt.Sprintf("%-40s %-27s %-10d %-15s %-20s",
			name, formattedStatus, restarts, ready, age)

		if i == m.cursor {
			b.WriteString(SelectedItemStyle.Render(row))
		} else {
			b.WriteString(UnselectedItemStyle.Render(row))
		}
		b.WriteString("\n")
	}

	return b.String()
}

// getPodStatus returns the pod status
func (m *PodsModel) getPodStatus(pod *corev1.Pod) string {
	status := string(pod.Status.Phase)

	// Check if pod is running but containers are not ready
	if pod.Status.Phase == corev1.PodRunning {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status != corev1.ConditionTrue {
				status = "NotReady"
				break
			}
		}
	}

	// Check for container statuses
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.State.Waiting != nil {
			status = containerStatus.State.Waiting.Reason
			break
		}
		if containerStatus.State.Terminated != nil {
			status = containerStatus.State.Terminated.Reason
			break
		}
	}

	return Truncate(status, 15)
}

// getPodRestarts returns the total number of restarts for a pod
func (m *PodsModel) getPodRestarts(pod *corev1.Pod) int {
	restarts := 0
	for _, containerStatus := range pod.Status.ContainerStatuses {
		restarts += int(containerStatus.RestartCount)
	}
	return restarts
}

// getPodReady returns the ready status of a pod
func (m *PodsModel) getPodReady(pod *corev1.Pod) string {
	readyContainers := 0
	totalContainers := len(pod.Status.ContainerStatuses)

	for _, containerStatus := range pod.Status.ContainerStatuses {
		if containerStatus.Ready {
			readyContainers++
		}
	}

	return fmt.Sprintf("%d/%d", readyContainers, totalContainers)
}

// formatAge formats a timestamp to relative time
func (m *PodsModel) formatAge(t time.Time) string {
	duration := time.Since(t.UTC())

	switch {
	case duration < time.Minute:
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	case duration < time.Hour:
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	case duration < 24*time.Hour:
		return fmt.Sprintf("%dh", int(duration.Hours()))
	default:
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

// SetLoading sets the loading state
func (m *PodsModel) SetLoading(loading bool) {
	m.loading = loading
}

// SetPods sets the pods data
func (m *PodsModel) SetPods(releaseName string, pods []corev1.Pod) {
	m.releaseName = releaseName
	m.pods = pods
	m.loading = false
	// Reset cursor if out of bounds
	if m.cursor >= len(m.pods) && len(m.pods) > 0 {
		m.cursor = len(m.pods) - 1
	}
}

// MoveCursorUp moves the cursor up
func (m *PodsModel) MoveCursorUp() {
	if m.cursor > 0 {
		m.cursor--
	}
}

// MoveCursorDown moves the cursor down
func (m *PodsModel) MoveCursorDown() {
	if m.cursor < len(m.pods)-1 {
		m.cursor++
	}
}

// GetSelectedPod returns the currently selected pod
func (m *PodsModel) GetSelectedPod() *corev1.Pod {
	if len(m.pods) == 0 || m.cursor >= len(m.pods) {
		return nil
	}
	return &m.pods[m.cursor]
}

// Reset resets the pods model
func (m *PodsModel) Reset() {
	m.releaseName = ""
	m.pods = nil
	m.cursor = 0
	m.loading = false
	m.err = nil
}
