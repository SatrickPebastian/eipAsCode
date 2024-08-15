package plugins

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// DockerComposePlugin is responsible for handling Docker Compose related tasks
type DockerComposePlugin struct{}

// Execute runs the Docker Compose plugin
func (p *DockerComposePlugin) Execute() error {
	// Define the path to the Docker Compose model file
	dockerComposeModelPath := filepath.Join("docker-compose.yaml")

	// Check if the file exists
	if _, err := exec.Command("test", "-f", dockerComposeModelPath).Output(); err != nil {
		return fmt.Errorf("docker-compose.yaml file not found: %w", err)
	}

	// Apply the Docker Compose model using docker-compose
	cmd := exec.Command("docker-compose", "-f", dockerComposeModelPath, "up", "-d")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply docker-compose.yaml: %w, output: %s", err, string(output))
	}

	fmt.Printf("Successfully applied Docker Compose model: %s\n", string(output))
	return nil
}

// Destroy stops and removes the Docker Compose services
func (p *DockerComposePlugin) Destroy() error {
	// Define the path to the Docker Compose model file
	dockerComposeModelPath := filepath.Join("docker-compose.yaml")

	// Check if the file exists
	if _, err := exec.Command("test", "-f", dockerComposeModelPath).Output(); err != nil {
		return fmt.Errorf("docker-compose.yaml file not found: %w", err)
	}

	// Stop and remove the Docker Compose services using docker-compose down
	cmd := exec.Command("docker-compose", "-f", dockerComposeModelPath, "down")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Handle cases where the services or containers might not exist
		if strings.Contains(string(output), "No such service") || strings.Contains(string(output), "No containers to remove") {
			fmt.Printf("No services or containers to remove: %s\n", string(output))
		} else {
			return fmt.Errorf("failed to stop and remove Docker Compose services: %w, output: %s", err, string(output))
		}
	} else {
		fmt.Printf("Successfully stopped and removed Docker Compose services: %s\n", string(output))
	}

	return nil
}
