package plugins

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type DockerComposePlugin struct{}

func (p *DockerComposePlugin) Execute() error {
	dockerComposeModelPath := filepath.Join("docker-compose.yaml")

	if _, err := exec.Command("test", "-f", dockerComposeModelPath).Output(); err != nil {
		return fmt.Errorf("docker-compose.yaml file not found: %w", err)
	}

	cmd := exec.Command("docker-compose", "-f", dockerComposeModelPath, "up", "-d", "--pull", "always")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply docker-compose.yaml: %w, output: %s", err, string(output))
	}

	fmt.Printf("Successfully applied Docker Compose model: %s\n", string(output))
	return nil
}

func (p *DockerComposePlugin) Destroy() error {
	dockerComposeModelPath := filepath.Join("docker-compose.yaml")

	if _, err := os.Stat(dockerComposeModelPath); os.IsNotExist(err) {
		fmt.Println("docker-compose.yaml file not found. Skipping destruction process.")
		return nil
	}

	cmd := exec.Command("docker-compose", "-f", dockerComposeModelPath, "down", "--rmi", "all", "--volumes", "--remove-orphans")
	output, err := cmd.CombinedOutput()
	if err != nil {
		//handles cases where the services or containers might not exist
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
