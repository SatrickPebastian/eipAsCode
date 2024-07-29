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
	volumes := map[string]interface{}{}

	for _, filter := range model.Filters {
		host := utils.FindHostByName(model.Hosts.FilterHosts, filter.Host)
		if host != nil && host.Type == "DockerCompose" {
			image := utils.FindArtifactImage(model.DeploymentArtifacts, filter.Artifact)
			service, serviceVolumes := createDockerComposeService(model, filter, image)
			serviceName := utils.SanitizeName(filter.Name)
			services[serviceName] = service
			for _, volume := range serviceVolumes {
				volumes[volume] = map[string]interface{}{}
			}
		}
	}

	composeFile := map[string]interface{}{
		"version":  "3",
		"services": services,
		"volumes":  volumes,
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

func createDockerComposeService(model *models.Model, filter models.Filter, image string) (map[string]interface{}, []string) {
	envVars := []string{}
	volumes := []string{}
	volumeMounts := []string{}

	for _, mapping := range filter.Mappings {
		parts := strings.Split(mapping, ":")
		if len(parts) == 2 {
			pipeName := parts[1]
			pipe := utils.FindQueueByName(model.Pipes.Queues, pipeName)
			if pipe != nil {
				pipeHost := utils.FindHostByName(model.Hosts.PipeHosts, pipe.Host)
				if pipeHost != nil {
					value := fmt.Sprintf("%s://%s:%s@%s:%s,%s",
						pipe.Protocol,
						pipeHost.AdditionalProps["username"],
						pipeHost.AdditionalProps["password"],
						pipeHost.AdditionalProps["host_address"],
						pipeHost.AdditionalProps["messaging_port"],
						pipe.Name, // add the pipe name at the end
					)
					envVars = append(envVars, fmt.Sprintf("%s=%s", parts[0], value))
				}
			}
		}
	}

	// Add environment variables for filter type configs
	filterType := utils.FindFilterTypeByName(model.FilterTypes, filter.Type)
	if filterType != nil {
		for _, config := range filterType.Configs {
			value, exists := filter.AdditionalProps[config.Name]
			if !exists {
				value = fmt.Sprintf("%v", config.Default)
			}
			if config.File {
				// Handle file-based config
				volumeName := strings.ToLower(filter.Name + "-" + config.Name)
				volumes = append(volumes, volumeName)
				volumeMounts = append(volumeMounts, fmt.Sprintf("%s:/etc/config/%s", volumeName, config.Name))

				// Update the value to point to the new mount path
				value = fmt.Sprintf("/etc/config/%s", config.Name)
			}

			envVars = append(envVars, fmt.Sprintf("%s=%s", config.Name, utils.ConvertToProperType(value)))
		}
	}

	service := map[string]interface{}{
		"image":       image,
		"environment": envVars,
		"volumes":     volumeMounts,
	}

	return service, volumes
}
