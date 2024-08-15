package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings" // <-- This import was missing
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

	// Start capturing live output from the command
	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr

	err = applyCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Terraform apply: %w", err)
	}

	// Wait for the command to finish
	err = applyCmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to apply Terraform configuration: %w", err)
	}

	fmt.Printf("Successfully applied Terraform model.\n")

	return nil
}

// Destroy removes the Terraform managed resources and cleans up any state and lock files
func (p *TerraformPlugin) Destroy() error {
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

	// Destroy the Terraform managed resources
	destroyCmd := exec.Command("terraform", "destroy", "-auto-approve")
	destroyOutput, err := destroyCmd.CombinedOutput()
	if err != nil {
		// Handle cases where there are no resources to destroy
		if strings.Contains(string(destroyOutput), "No changes. Infrastructure is up-to-date.") || strings.Contains(string(destroyOutput), "No resources found") {
			fmt.Printf("No resources to destroy: %s\n", string(destroyOutput))
		} else {
			return fmt.Errorf("failed to destroy Terraform managed resources: %w, output: %s", err, string(destroyOutput))
		}
	} else {
		fmt.Printf("Successfully destroyed Terraform managed resources: %s\n", string(destroyOutput))
	}

	// Clean up any Terraform state files to prevent future locking issues
	fmt.Println("Cleaning up Terraform state and lock files...")
	stateFiles := []string{"terraform.tfstate", "terraform.tfstate.backup", "terraform.lock.hcl"}
	for _, stateFile := range stateFiles {
		if _, err := os.Stat(stateFile); err == nil {
			err := os.Remove(stateFile)
			if err != nil {
				return fmt.Errorf("failed to remove %s: %w", stateFile, err)
			}
			fmt.Printf("Removed %s\n", stateFile)
		}
	}

	fmt.Println("Terraform cleanup complete.")
	return nil
}
