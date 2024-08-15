package plugins

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// KubernetesPlugin is responsible for handling Kubernetes related tasks
type KubernetesPlugin struct{}

// Execute runs the Kubernetes plugin
func (p *KubernetesPlugin) Execute() error {
	// Define the path to the Kubernetes model file
	kubernetesModelPath := filepath.Join(".", "kubernetesModel.yaml")

	// Check if the file exists
	if _, err := exec.Command("test", "-f", kubernetesModelPath).Output(); err != nil {
		return fmt.Errorf("kubernetesModel.yaml file not found: %w", err)
	}

	// Apply the Kubernetes model using kubectl
	cmd := exec.Command("kubectl", "apply", "-f", kubernetesModelPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply kubernetesModel.yaml: %w, output: %s", err, string(output))
	}

	// Check the status of the deployment
	// Assuming the deployments have the same name as the filters
	checkCmd := exec.Command("kubectl", "get", "deployments")
	checkOutput, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get deployment status: %w, output: %s", err, string(checkOutput))
	}

	fmt.Printf("Successfully applied Kubernetes model: %s\n", string(output))
	return nil
}

// Destroy removes the Kubernetes resources defined in the kubernetesModel.yaml file
func (p *KubernetesPlugin) Destroy() error {
	// Define the path to the Kubernetes model file
	kubernetesModelPath := filepath.Join(".", "kubernetesModel.yaml")

	// Check if the file exists
	if _, err := exec.Command("test", "-f", kubernetesModelPath).Output(); err != nil {
		return fmt.Errorf("kubernetesModel.yaml file not found: %w", err)
	}

	// Delete the Kubernetes resources using kubectl
	cmd := exec.Command("kubectl", "delete", "-f", kubernetesModelPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Handle "NotFound" errors gracefully
		if strings.Contains(string(output), "NotFound") {
			fmt.Printf("Some resources were not found: %s\n", string(output))
		} else {
			return fmt.Errorf("failed to delete Kubernetes resources: %w, output: %s", err, string(output))
		}
	} else {
		fmt.Printf("Successfully deleted Kubernetes resources: %s\n", string(output))
	}

	return nil
}
