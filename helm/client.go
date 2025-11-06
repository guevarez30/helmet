package helm

import (
	"context"
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/repo"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Client wraps the Helm SDK for simplified operations
type Client struct {
	settings  *cli.EnvSettings
	actionCfg *action.Configuration
	namespace string
	k8sGetter genericclioptions.RESTClientGetter
}

// NewClient creates a new Helm client
func NewClient(namespace string) (*Client, error) {
	settings := cli.New()

	if namespace != "" {
		settings.SetNamespace(namespace)
	}

	actionCfg := new(action.Configuration)

	// Initialize action configuration with current kubeconfig context
	configFlags := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}

	if err := actionCfg.Init(configFlags, settings.Namespace(), os.Getenv("HELM_DRIVER"), func(format string, v ...interface{}) {
		// Debug logging (can be disabled in production)
		// fmt.Printf(format, v...)
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize helm client: %w", err)
	}

	return &Client{
		settings:  settings,
		actionCfg: actionCfg,
		namespace: settings.Namespace(),
		k8sGetter: configFlags,
	}, nil
}

// ListReleases returns all releases in the current namespace
func (c *Client) ListReleases(all bool) ([]*release.Release, error) {
	client := action.NewList(c.actionCfg)

	if all {
		client.All = true
		client.AllNamespaces = false
	}

	client.SetStateMask()

	releases, err := client.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to list releases: %w", err)
	}

	return releases, nil
}

// InstallChart installs a new chart
func (c *Client) InstallChart(name, chart string, vals map[string]interface{}) (*release.Release, error) {
	client := action.NewInstall(c.actionCfg)
	client.Namespace = c.namespace
	client.ReleaseName = name

	chartPath, err := client.ChartPathOptions.LocateChart(chart, c.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	chartRequested, err := c.loadChart(chartPath)
	if err != nil {
		return nil, err
	}

	rel, err := client.Run(chartRequested, vals)
	if err != nil {
		return nil, fmt.Errorf("failed to install chart: %w", err)
	}

	return rel, nil
}

// UpgradeRelease upgrades an existing release
func (c *Client) UpgradeRelease(name, chart string, vals map[string]interface{}) (*release.Release, error) {
	client := action.NewUpgrade(c.actionCfg)
	client.Namespace = c.namespace

	chartPath, err := client.ChartPathOptions.LocateChart(chart, c.settings)
	if err != nil {
		return nil, fmt.Errorf("failed to locate chart: %w", err)
	}

	chartRequested, err := c.loadChart(chartPath)
	if err != nil {
		return nil, err
	}

	rel, err := client.Run(name, chartRequested, vals)
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade release: %w", err)
	}

	return rel, nil
}

// UninstallRelease uninstalls a release
func (c *Client) UninstallRelease(name string) (*release.UninstallReleaseResponse, error) {
	client := action.NewUninstall(c.actionCfg)

	resp, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to uninstall release: %w", err)
	}

	return resp, nil
}

// GetReleaseValues returns the values for a release
func (c *Client) GetReleaseValues(name string) (map[string]interface{}, error) {
	client := action.NewGetValues(c.actionCfg)
	client.AllValues = true

	vals, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release values: %w", err)
	}

	return vals, nil
}

// RollbackRelease rolls back a release to a previous revision
func (c *Client) RollbackRelease(name string, revision int) error {
	client := action.NewRollback(c.actionCfg)
	client.Version = revision

	if err := client.Run(name); err != nil {
		return fmt.Errorf("failed to rollback release: %w", err)
	}

	return nil
}

// GetReleaseHistory returns the revision history for a release
func (c *Client) GetReleaseHistory(name string) ([]*release.Release, error) {
	client := action.NewHistory(c.actionCfg)
	client.Max = 256 // Maximum number of revisions to retrieve

	releases, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release history: %w", err)
	}

	return releases, nil
}

// GetRelease returns detailed information about a release
func (c *Client) GetRelease(name string) (*release.Release, error) {
	client := action.NewGet(c.actionCfg)

	rel, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release: %w", err)
	}

	return rel, nil
}

// GetReleaseStatus returns the status of a release
func (c *Client) GetReleaseStatus(name string) (*release.Release, error) {
	client := action.NewStatus(c.actionCfg)

	rel, err := client.Run(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get release status: %w", err)
	}

	return rel, nil
}

// ListRepositories returns all configured Helm repositories
func (c *Client) ListRepositories() ([]*repo.Entry, error) {
	repoFile := c.settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load repository file: %w", err)
	}

	return f.Repositories, nil
}

// AddRepository adds a new Helm repository
func (c *Client) AddRepository(name, url string) error {
	repoFile := c.settings.RepositoryConfig

	entry := &repo.Entry{
		Name: name,
		URL:  url,
	}

	chartRepo, err := repo.NewChartRepository(entry, getter.All(c.settings))
	if err != nil {
		return fmt.Errorf("failed to create chart repository: %w", err)
	}

	if _, err := chartRepo.DownloadIndexFile(); err != nil {
		return fmt.Errorf("failed to download index file: %w", err)
	}

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository file: %w", err)
	}

	f.Update(entry)

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	return nil
}

// RemoveRepository removes a Helm repository
func (c *Client) RemoveRepository(name string) error {
	repoFile := c.settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository file: %w", err)
	}

	if !f.Remove(name) {
		return fmt.Errorf("repository %s not found", name)
	}

	if err := f.WriteFile(repoFile, 0644); err != nil {
		return fmt.Errorf("failed to write repository file: %w", err)
	}

	return nil
}

// UpdateRepositories updates all configured repositories
func (c *Client) UpdateRepositories(ctx context.Context) error {
	repoFile := c.settings.RepositoryConfig

	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return fmt.Errorf("failed to load repository file: %w", err)
	}

	for _, entry := range f.Repositories {
		chartRepo, err := repo.NewChartRepository(entry, getter.All(c.settings))
		if err != nil {
			return fmt.Errorf("failed to create chart repository: %w", err)
		}

		if _, err := chartRepo.DownloadIndexFile(); err != nil {
			return fmt.Errorf("failed to update repository %s: %w", entry.Name, err)
		}
	}

	return nil
}

// loadChart is a helper function to load a chart from a path
func (c *Client) loadChart(path string) (*chart.Chart, error) {
	return loader.Load(path)
}

// GetChartVersions fetches available versions for a chart
func (c *Client) GetChartVersions(chartName string) ([]*repo.ChartVersion, error) {
	// Load repository file
	repoFile := c.settings.RepositoryConfig
	f, err := repo.LoadFile(repoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load repository file: %w", err)
	}

	// Find the chart in repositories
	var versions []*repo.ChartVersion
	for _, entry := range f.Repositories {
		chartRepo, err := repo.NewChartRepository(entry, getter.All(c.settings))
		if err != nil {
			continue
		}

		// Load the repository index
		indexFile, err := repo.LoadIndexFile(chartRepo.CachePath)
		if err != nil {
			continue
		}

		// Look for the chart (try with and without repo prefix)
		chartKey := chartName
		if parts := splitChartName(chartName); len(parts) == 2 {
			chartKey = parts[1]
		}

		if chartVersions, ok := indexFile.Entries[chartKey]; ok {
			versions = append(versions, chartVersions...)
		}
	}

	if len(versions) == 0 {
		return nil, fmt.Errorf("chart %s not found in any repository", chartName)
	}

	return versions, nil
}

// splitChartName splits repo/chart into parts
func splitChartName(name string) []string {
	for i, c := range name {
		if c == '/' {
			return []string{name[:i], name[i+1:]}
		}
	}
	return []string{name}
}

// SetNamespace changes the active namespace for operations
func (c *Client) SetNamespace(namespace string) {
	c.namespace = namespace
	c.settings.SetNamespace(namespace)
}

// GetNamespace returns the current namespace
func (c *Client) GetNamespace() string {
	return c.namespace
}

// GetReleasePods returns pods associated with a Helm release
func (c *Client) GetReleasePods(releaseName string) ([]corev1.Pod, error) {
	// Get the kubernetes clientset
	restConfig, err := c.k8sGetter.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get REST config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	// Helm uses the label "app.kubernetes.io/instance=<release-name>" for resources
	labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName)

	pods, err := clientset.CoreV1().Pods(c.namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return pods.Items, nil
}
