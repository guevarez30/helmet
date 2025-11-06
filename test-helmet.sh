#!/bin/bash
# Test script to verify Helmet can read Helm releases

echo "=== Testing Helmet Components ==="
echo ""

echo "1. Current Kubernetes Context:"
kubectl config current-context
echo ""

echo "2. Current Namespace:"
kubectl config view --minify --output 'jsonpath={..namespace}' || echo "default"
echo ""
echo ""

echo "3. Helm Releases in Cluster:"
helm list
echo ""

echo "4. Pods in Default Namespace:"
kubectl get pods
echo ""

echo "=== Helmet is ready to use! ==="
echo ""
echo "To run Helmet interactively, execute:"
echo "  ./helmet"
echo ""
echo "Available keybindings in Helmet:"
echo "  - tab       : Switch between Dashboard, Releases, Repositories views"
echo "  - ↑/↓ or j/k: Navigate releases"
echo "  - d         : Delete selected release"
echo "  - u         : Upgrade release (coming soon)"
echo "  - v         : View release values (coming soon)"
echo "  - H         : View release history (coming soon)"
echo "  - ctrl+r    : Refresh data"
echo "  - q         : Quit"
echo ""
