package transformators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"eicoda/models"
	"eicoda/utils"
	"gopkg.in/yaml.v2"
)

// DockerComposeTransformator is responsible for transforming the model to Docker Compose format
type DockerComposeTransformator struct{}

// Transform transforms the model to Docker Compose format
func (t *DockerComposeTransformator) Transform(model *models.Model) error {
	services := make(map[string]interface{})

	for _, filter := range model.Filters {
		host := utils.FindHostByName(model.Hosts.FilterHosts, filter.Host)
		if host != nil && host.Type == "DockerCompose" {
			service := createDockerComposeService(filter, model.DeploymentArtifacts.Image)
			serviceName := utils.SanitizeName(filter.Name)
			services[serviceName] = service
		}
	}

	composeFile := map[string]interface{}{
		"version":  "3",
		"services": services,
	}

	// Generate the file at the project root
	outputPath := filepath.Join(".", "docker-compose.yaml")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create Docker Compose model file: %w", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	err = encoder.Encode(composeFile)
	if err != nil {
		return fmt.Errorf("failed to write Docker Compose model to file: %w", err)
	}
	encoder.Close()

	fmt.Printf("Successfully created Docker Compose model file at %s\n", outputPath)
	return nil
}

func createDockerComposeService(filter models.Filter, image string) map[string]interface{} {
	envVars := []string{}
	for _, mapping := range filter.Mappings {
		parts := strings.Split(mapping, ":")
		if len(parts) == 2 {
			envVars = append(envVars, fmt.Sprintf("%s=%s", parts[0], parts[1]))
		}
	}

	service := map[string]interface{}{
		"image":       image,
		"environment": envVars,
	}

	return service
}
