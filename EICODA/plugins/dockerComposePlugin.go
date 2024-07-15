package plugins

import (
	"fmt"
	"os/exec"
	"path/filepath"
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