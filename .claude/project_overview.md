# Helmet - Project Overview

## Project Summary
Helmet is a modern, interactive terminal UI (TUI) for managing Kubernetes clusters through Helm. Built with Go and the Bubble Tea framework, it provides a beautiful, keyboard-driven interface for Helm operations, release management, and cluster monitoring.

## Project Goals
- Create a fast, responsive Helm/Kubernetes management tool
- Provide a modern, visually appealing terminal interface
- Support vim-style keybindings for power users
- Offer real-time release and cluster management
- Simplify common Helm workflows

## Technology Stack

### Core Frameworks
- **Go 1.24+** - Primary language
- **Bubble Tea** - TUI framework using The Elm Architecture (Model-View-Update)
- **Lipgloss** - Terminal styling library for colors and layouts
- **Bubbles** - Pre-built TUI components (viewport, textinput, keys)

### Helm/Kubernetes Integration
- **Helm SDK** (`helm.sh/helm/v3/pkg/*`) - Official Helm Go SDK
- **Kubernetes Client** (`k8s.io/client-go`) - K8s API access for context/namespace management

## Key Design Decisions

### Architecture Pattern
- **MVC-like structure** with Bubble Tea's Model-View-Update pattern
- Separation of concerns:
  - `helm/` - Helm SDK wrapper/client
  - `kubernetes/` - K8s context and namespace management
  - `ui/` - Bubble Tea models and views

### Visual Design
- **Color Scheme**: Kubernetes Blue (#326CE5), Purple (#7D56F4), Cyan (#8BE9FD)
- **Typography**: Clean, tabular layouts with proper column alignment
- **Status Colors**: Green (deployed), Red (failed), Yellow (pending/warning)
- **Tab Navigation**: Active tab with K8s blue background, inactive tabs muted

### User Experience
- Vim-style keybindings (hjkl navigation)
- Tab-based view switching
- Visual feedback for actions (status messages, loading indicators)
- Context and namespace display in status bar
- Non-blocking operations with progress indicators

## Project Structure

```
helmet/
├── main.go              # Application entry point
├── helm/                # Helm client wrapper
│   └── client.go        # Simplified Helm SDK interface
├── kubernetes/          # Kubernetes integration
│   └── context.go       # Context and namespace management
├── ui/                  # Bubble Tea UI layer
│   ├── model.go         # Main application model & view routing
│   ├── dashboard.go     # Dashboard statistics view
│   ├── releases.go      # Release management view
│   ├── styles.go        # Lipgloss style definitions
│   └── keys.go          # Keybinding definitions
├── .claude/             # Documentation for AI assistants
└── test-helmet.sh       # Test script for local cluster
```

## Current State (2025-11-06)

### Implemented Features

#### Phase 1 MVP ✅
- Helm client wrapper with all major operations
- Kubernetes context and namespace management
- Dashboard with release statistics (deployed, failed, pending)
- Releases list view with detailed information
- Release deletion functionality
- Tab navigation between views
- Status bar with context/cluster/namespace display
- Visual feedback for operations
- Vim-style keybindings
- Kubernetes-inspired color scheme

#### Phase 2 Advanced Features ✅
- **Repositories Management**: Full CRUD operations
  - List all configured repositories
  - Add new repositories (HTTP, local file://, OCI)
  - Remove repositories
  - Update all repositories
- **Values Viewer**: Display and scroll through YAML values for any release
- **Install Charts**: Interactive installation form
  - Auto-discover local charts in current directory
  - Support for repository charts and local paths
  - Custom namespace selection
- **Chart Catalog**: Multi-select browser
  - 10+ popular public charts from multiple repos
  - Bitnami, Prometheus, Grafana, Ingress, Cert Manager, ArgoCD
  - Automatic repository setup
  - Batch installation
- **Quick Start**: One-click demo environment setup from Dashboard
- **Local Chart Discovery**: Automatically find Chart.yaml files in project
- **Modal Views**: Proper overlay UI for forms and viewers

### Test Environment
✅ kind cluster setup (`kind-helmet-test`)
✅ Sample releases:
  - `my-nginx` - NGINX web server (bitnami/nginx)
  - `my-redis` - Redis cache (bitnami/redis)

### Planned Features (Phase 3)
- [ ] Release upgrade functionality with version selection
- [ ] Release history viewer with diff
- [ ] Rollback to previous revisions
- [ ] Kubernetes resource viewer (pods, services, etc.)
- [ ] Pod logs viewer
- [ ] Multi-context switcher UI
- [ ] Namespace switcher
- [ ] Search and filter releases/charts
- [ ] Release testing integration
- [ ] Export/import configurations

## Development Context
This project was built collaboratively with Claude Code, with a focus on:
- Modern Go patterns and idioms
- Clean, maintainable code structure
- User experience and visual polish
- Incremental feature development
- Real-time feedback and iteration
