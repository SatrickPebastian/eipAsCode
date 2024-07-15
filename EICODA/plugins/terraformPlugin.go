package plugins

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

// TerraformPlugin is responsible for handling Terraform related tasks
type TerraformPlugin struct{}

// Execute runs the Terraform plugin
func (p *TerraformPlugin) Execute() error {
	// Define the path to the Terraform model file
	terraformModelPath := filepath.Join(".", "rabbitMqModel.tf")

	// Check if the file exists
	if _, err := exec.Command("test", "-f", terraformModelPath).Output(); err != nil {
		return fmt.Errorf("rabbitMqModel.tf file not found: %w", err)
	}

	// Initialize Terraform with the correct provider source
	initCmd := exec.Command("terraform", "init")
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Terraform: %w, output: %s", err, string(initOutput))
	}

	// Apply the Terraform configuration
	applyCmd := exec.Command("terraform", "apply", "-auto-approve")
	applyOutput, err := applyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply Terraform configuration: %w, output: %s", err, string(applyOutput))
	}

	fmt.Printf("Successfully applied Terraform model: %s\n", string(applyOutput))
	return nil
}
