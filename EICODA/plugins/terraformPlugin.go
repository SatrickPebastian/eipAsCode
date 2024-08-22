package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type TerraformPlugin struct{}

func (p *TerraformPlugin) Execute() error {
	terraformModelPath := filepath.Join(".", "rabbitMqModel.tf")

	if _, err := exec.Command("test", "-f", terraformModelPath).Output(); err != nil {
		return fmt.Errorf("rabbitMqModel.tf file not found: %w", err)
	}

	initCmd := exec.Command("terraform", "init")
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Terraform: %w, output: %s", err, string(initOutput))
	}
	applyCmd := exec.Command("terraform", "apply", "-auto-approve")

	applyCmd.Stdout = os.Stdout
	applyCmd.Stderr = os.Stderr

	err = applyCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start Terraform apply: %w", err)
	}

	err = applyCmd.Wait()
	if err != nil {
		return fmt.Errorf("failed to apply Terraform configuration: %w", err)
	}

	fmt.Printf("Successfully applied Terraform model.\n")

	return nil
}

func (p *TerraformPlugin) Destroy() error {
	terraformModelPath := filepath.Join(".", "rabbitMqModel.tf")

	if _, err := exec.Command("test", "-f", terraformModelPath).Output(); err != nil {
		return fmt.Errorf("rabbitMqModel.tf file not found: %w", err)
	}

	initCmd := exec.Command("terraform", "init")
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Terraform: %w, output: %s", err, string(initOutput))
	}

	destroyCmd := exec.Command("terraform", "destroy", "-auto-approve")
	destroyOutput, err := destroyCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(destroyOutput), "No changes. Infrastructure is up-to-date.") || strings.Contains(string(destroyOutput), "No resources found") {
			fmt.Printf("No resources to destroy: %s\n", string(destroyOutput))
		} else {
			return fmt.Errorf("failed to destroy Terraform managed resources: %w, output: %s", err, string(destroyOutput))
		}
	} else {
		fmt.Printf("Successfully destroyed Terraform managed resources: %s\n", string(destroyOutput))
	}

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
