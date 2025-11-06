# Helmet 🪖

A modern, interactive terminal UI for managing Kubernetes clusters through Helm.

![Kubernetes](https://img.shields.io/badge/kubernetes-%23326ce5.svg?style=for-the-badge&logo=kubernetes&logoColor=white)
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Helm](https://img.shields.io/badge/helm-0F1689?style=for-the-badge&logo=helm&logoColor=white)

## Features

### ✨ Core Functionality
- **Dashboard**: Real-time overview of all Helm releases (deployed, failed, pending)
- **Release Management**: List, install, delete, and inspect Helm releases
- **Repository Management**: Add, remove, and update Helm repositories (HTTP, local, OCI)
- **Values Viewer**: Display and scroll through YAML values for any release
- **Chart Catalog**: Browse and install popular charts from multiple repositories
- **Local Chart Discovery**: Automatically detect Helm charts in your project
- **Vim-style Navigation**: Keyboard-driven interface for power users

### 🎯 Highlights
- **Zero Configuration**: Uses your existing kubeconfig automatically
- **Quick Start**: One-click setup with example charts for testing
- **Local Repositories**: Support for `file://` paths to work with local charts
- **Multi-select Install**: Install multiple charts at once from the catalog
- **Beautiful UI**: Kubernetes-inspired color scheme with visual status indicators

## Installation

### Prerequisites
- Go 1.24 or later
- kubectl configured with cluster access
- Helm 3.x (optional, Helmet uses the SDK directly)

### Build from Source
```bash
git clone https://github.com/guevarez30/helmet.git
cd helmet
go build -buildvcs=false -o helmet
./helmet
```

## Quick Start

### 1. Start Helmet
```bash
./helmet
```

### 2. Set Up Test Environment
- Press `s` from the Dashboard to open the Chart Catalog
- Use `↑/↓` or `j/k` to navigate charts
- Press `space` to select charts you want to install
- Press `enter` to install selected charts

### 3. Explore Features
- Press `tab` to cycle through views (Dashboard → Releases → Repositories)
- Press `v` on a release to view its YAML values
- Press `i` to install a custom chart
- Press `a` in Repositories to add a new repo

## Usage

### Navigation

#### Global Keybindings
| Key | Action |
|-----|--------|
| `tab` | Switch views |
| `q` | Quit |
| `ctrl+r` | Refresh current view |

#### Dashboard View
| Key | Action |
|-----|--------|
| `s` | Quick start (open chart catalog) |
| `i` | Install chart |

#### Releases View
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate releases |
| `i` | Install new chart |
| `d` | Delete selected release |
| `v` | View release values (YAML) |

#### Repositories View
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate repositories |
| `a` | Add repository |
| `r` | Remove selected repository |
| `U` | Update all repositories |

#### Chart Catalog
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate charts |
| `space` | Select/deselect chart |
| `enter` | Install selected charts |
| `esc` | Cancel |

### Installing Charts

#### From Repository
1. Press `i` from Dashboard or Releases view
2. Enter release name (e.g., `my-app`)
3. Enter chart path (e.g., `bitnami/nginx`)
4. Enter namespace (default: `default`)
5. Press `enter` to install

#### From Local Directory
1. Press `i` to open install form
2. Enter release name
3. Enter local path (e.g., `./mychart` or `/path/to/chart`)
4. The form automatically discovers charts in your current directory
5. Press `enter` to install

### Adding Repositories

#### Public HTTP Repository
```
Repository Name: bitnami
Repository URL:  https://charts.bitnami.com/bitnami
```

#### Local File Repository
```
Repository Name: my-charts
Repository URL:  file:///Users/you/helm-charts
```

#### OCI Registry
```
Repository Name: azure
Repository URL:  oci://mcr.microsoft.com/helm
```

## Architecture

```
helmet/
├── main.go              # Application entry point
├── helm/
│   ├── client.go        # Helm SDK wrapper
│   └── discovery.go     # Local chart discovery
├── kubernetes/
│   └── context.go       # Kubeconfig management
└── ui/
    ├── model.go         # Main application model (Bubble Tea)
    ├── dashboard.go     # Dashboard view
    ├── releases.go      # Releases list view
    ├── repositories.go  # Repositories view
    ├── values.go        # YAML values viewer
    ├── install.go       # Chart installation form
    ├── catalog.go       # Chart catalog browser
    ├── addrepo.go       # Add repository form
    ├── styles.go        # UI styling
    └── keys.go          # Keybindings
```

### Built With
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** - TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** - Terminal styling
- **[Helm SDK](https://helm.sh/)** - Helm operations
- **[client-go](https://github.com/kubernetes/client-go)** - Kubernetes API client

## Examples

### Test with kind (Kubernetes in Docker)
```bash
# Create a test cluster
kind create cluster --name helmet-test

# Start Helmet
./helmet

# Press 's' to open chart catalog
# Select charts and press enter to install
```

### Working with Local Charts
```bash
# Navigate to a directory with Helm charts
cd my-project

# Start Helmet
../helmet

# Press 'i' to install
# Local charts will be auto-discovered and listed
```

### Managing Multiple Clusters
```bash
# Helmet uses your current kubectl context
kubectl config get-contexts
kubectl config use-context my-cluster

# Start Helmet
./helmet
```

## Development

### Building
```bash
go build -buildvcs=false -o helmet
```

### Dependencies
```bash
go mod tidy
go mod download
```

### Testing
```bash
# Create test cluster
kind create cluster --name helmet-test

# Run test script
./test-helmet.sh
```

## Roadmap

### Phase 3 (Planned)
- [ ] Release upgrade with version selection
- [ ] Release history and rollback
- [ ] Kubernetes resource viewer
- [ ] Pod logs viewer
- [ ] Multi-context switcher
- [ ] Namespace switcher
- [ ] Search and filter functionality

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details

## Acknowledgments

- Built with [Claude Code](https://claude.com/claude-code)
- Kubernetes community for excellent tooling
- Helm project for the comprehensive SDK

## Support

- 📖 [Documentation](./.claude/)
- 🐛 [Issue Tracker](https://github.com/guevarez30/helmet/issues)
- 💬 Discussions welcome in issues

---

**Helmet** - Manage your Kubernetes clusters with style 🪖
