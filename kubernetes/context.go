package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// ContextManager handles Kubernetes context and namespace operations
type ContextManager struct {
	config       *api.Config
	configPath   string
	currentCtx   string
	currentNs    string
	clientset    *kubernetes.Clientset
}

// NewContextManager creates a new context manager
func NewContextManager() (*ContextManager, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configPath := loadingRules.GetDefaultFilename()

	config, err := clientcmd.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	cm := &ContextManager{
		config:     config,
		configPath: configPath,
		currentCtx: config.CurrentContext,
	}

	// Get current namespace
	if ctx, ok := config.Contexts[config.CurrentContext]; ok {
		if ctx.Namespace != "" {
			cm.currentNs = ctx.Namespace
		} else {
			cm.currentNs = "default"
		}
	}

	// Initialize Kubernetes clientset
	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}
	cm.clientset = clientset

	return cm, nil
}

// ListContexts returns all available contexts
func (cm *ContextManager) ListContexts() []string {
	contexts := make([]string, 0, len(cm.config.Contexts))
	for name := range cm.config.Contexts {
		contexts = append(contexts, name)
	}
	return contexts
}

// GetCurrentContext returns the current context name
func (cm *ContextManager) GetCurrentContext() string {
	return cm.currentCtx
}

// SwitchContext changes the current context
func (cm *ContextManager) SwitchContext(ctx string) error {
	if _, ok := cm.config.Contexts[ctx]; !ok {
		return fmt.Errorf("context %s not found", ctx)
	}

	cm.config.CurrentContext = ctx
	cm.currentCtx = ctx

	// Update namespace from new context
	if context, ok := cm.config.Contexts[ctx]; ok {
		if context.Namespace != "" {
			cm.currentNs = context.Namespace
		} else {
			cm.currentNs = "default"
		}
	}

	// Persist the change
	if err := clientcmd.WriteToFile(*cm.config, cm.configPath); err != nil {
		return fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	return nil
}

// GetCurrentNamespace returns the current namespace
func (cm *ContextManager) GetCurrentNamespace() string {
	return cm.currentNs
}

// SetNamespace changes the namespace for the current context
func (cm *ContextManager) SetNamespace(ns string) error {
	ctx, ok := cm.config.Contexts[cm.currentCtx]
	if !ok {
		return fmt.Errorf("current context %s not found", cm.currentCtx)
	}

	ctx.Namespace = ns
	cm.currentNs = ns

	// Persist the change
	if err := clientcmd.WriteToFile(*cm.config, cm.configPath); err != nil {
		return fmt.Errorf("failed to write kubeconfig: %w", err)
	}

	return nil
}

// GetContextCluster returns the cluster name for a given context
func (cm *ContextManager) GetContextCluster(contextName string) string {
	if ctx, ok := cm.config.Contexts[contextName]; ok {
		return ctx.Cluster
	}
	return ""
}

// GetClusterServer returns the server URL for a given cluster
func (cm *ContextManager) GetClusterServer(clusterName string) string {
	if cluster, ok := cm.config.Clusters[clusterName]; ok {
		return cluster.Server
	}
	return ""
}

// GetContextInfo returns detailed information about the current context
func (cm *ContextManager) GetContextInfo() (cluster, server, namespace string) {
	cluster = cm.GetContextCluster(cm.currentCtx)
	server = cm.GetClusterServer(cluster)
	namespace = cm.currentNs
	return
}

// ListPods returns pods in the given namespace with optional label selector
func (cm *ContextManager) ListPods(namespace string, labelSelector string) ([]corev1.Pod, error) {
	pods, err := cm.clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	return pods.Items, nil
}

// GetPodsForRelease returns pods associated with a Helm release
func (cm *ContextManager) GetPodsForRelease(namespace, releaseName string) ([]corev1.Pod, error) {
	// Helm typically uses the label "app.kubernetes.io/instance=<release-name>"
	labelSelector := fmt.Sprintf("app.kubernetes.io/instance=%s", releaseName)
	return cm.ListPods(namespace, labelSelector)
}

// GetPodLogs returns logs for a specific pod
func (cm *ContextManager) GetPodLogs(namespace, podName string, tailLines int64) (string, error) {
	req := cm.clientset.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tailLines,
	})

	logs, err := req.DoRaw(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to get pod logs: %w", err)
	}

	return string(logs), nil
}
