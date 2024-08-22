package plugins

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type KubernetesPlugin struct{}

func (p *KubernetesPlugin) Execute() error {
	kubernetesModelPath := filepath.Join(".", "kubernetesModel.yaml")

	if _, err := exec.Command("test", "-f", kubernetesModelPath).Output(); err != nil {
		return fmt.Errorf("kubernetesModel.yaml file not found: %w", err)
	}

	cmd := exec.Command("kubectl", "apply", "-f", kubernetesModelPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply kubernetesModel.yaml: %w, output: %s", err, string(output))
	}

	// checks the status, assumes the deployments have same names as filters
	checkCmd := exec.Command("kubectl", "get", "deployments")
	checkOutput, err := checkCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get deployment status: %w, output: %s", err, string(checkOutput))
	}

	fmt.Printf("Successfully applied Kubernetes model: %s\n", string(output))
	return nil
}

func (p *KubernetesPlugin) Destroy() error {
	kubernetesModelPath := filepath.Join(".", "kubernetesModel.yaml")

	if _, err := exec.Command("test", "-f", kubernetesModelPath).Output(); err != nil {
		return fmt.Errorf("kubernetesModel.yaml file not found: %w", err)
	}

	cmd := exec.Command("kubectl", "delete", "-f", kubernetesModelPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
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
