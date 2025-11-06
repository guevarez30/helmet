package helm

import (
	"os"
	"path/filepath"
	"strings"
)

// LocalChart represents a discovered local Helm chart
type LocalChart struct {
	Name string
	Path string
}

// DiscoverLocalCharts finds Helm charts in the current directory and subdirectories
func DiscoverLocalCharts(searchPath string) ([]LocalChart, error) {
	var charts []LocalChart

	// Default to current directory if not specified
	if searchPath == "" {
		searchPath = "."
	}

	// Walk the directory tree looking for Chart.yaml files
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		// Skip hidden directories and node_modules, vendor, etc.
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") ||
			   name == "node_modules" ||
			   name == "vendor" ||
			   name == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if this is a Chart.yaml file
		if info.Name() == "Chart.yaml" {
			chartDir := filepath.Dir(path)
			chartName := filepath.Base(chartDir)

			charts = append(charts, LocalChart{
				Name: chartName,
				Path: chartDir,
			})
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return charts, nil
}
