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

// Transform transforms the model to Docker Compose format and optionally writes to a file
func (t *DockerComposeTransformator) Transform(model *models.Model, writeFile bool, baseDir string) (string, error) {
	services := make(map[string]interface{})
	volumes := map[string]interface{}{}

	for _, filter := range model.Filters {
		host := utils.FindHostByName(model.Hosts.FilterHosts, filter.Host)
		if host != nil && host.Type == "DockerEngine" {
			image := utils.FindArtifactImage(model.DeploymentArtifacts, filter.Artifact)
			service, serviceVolumes := createDockerComposeService(model, filter, image, baseDir)
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

	var sb strings.Builder
	encoder := yaml.NewEncoder(&sb)
	err := encoder.Encode(composeFile)
	if err != nil {
		return "", fmt.Errorf("failed to encode Docker Compose model: %w", err)
	}
	encoder.Close()

	// Write to file if writeFile is true
	if writeFile {
		outputPath := "docker-compose.yaml" // Write to the root folder
		err := os.WriteFile(outputPath, []byte(sb.String()), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write Docker Compose model to file: %w", err)
		}
	}

	return sb.String(), nil
}

func createDockerComposeService(model *models.Model, filter models.Filter, image string, baseDir string) (map[string]interface{}, []string) {
	envVars := []string{}
	volumes := []string{}
	volumeMounts := []string{}

	for _, mapping := range filter.Mappings {
		parts := strings.Split(mapping, ":")
		if len(parts) == 2 {
			pipeMapping := parts[1]
			var pipeName, routingKey string

			// Check if the mapping has a routing key defined with "->"
			if strings.Contains(pipeMapping, "->") {
				pipeParts := strings.Split(pipeMapping, "->")
				if len(pipeParts) == 2 {
					pipeName = pipeParts[0]
					routingKey = pipeParts[1]
				} else {
					pipeName = pipeMapping // Fallback in case of incorrect formatting
				}
			} else {
				pipeName = pipeMapping
			}

			var pipeType string
			var pipeHost *models.Host
			var pipeProtocol string

			// Check if it's a queue
			queue := utils.FindQueueByName(model.Pipes.Queues, pipeName)
			if queue != nil {
				pipeType = "queue"
				pipeHost = utils.FindHostByName(model.Hosts.PipeHosts, queue.Host)
				pipeProtocol = queue.Protocol
			} else {
				// If not a queue, it must be a topic
				topic := utils.FindTopicByName(model.Pipes.Topics, pipeName)
				if topic != nil {
					pipeType = "topic"
					pipeHost = utils.FindHostByName(model.Hosts.PipeHosts, topic.Host)
					pipeProtocol = topic.Protocol
				}
			}

			if pipeHost != nil {
				hostAddress := pipeHost.AdditionalProps["host_address"]
				if hostAddress == "localhost" {
					hostAddress = "host.docker.internal"
				}

				value := fmt.Sprintf("%s://%s:%s@%s:%s,%s,%s",
					pipeProtocol,
					pipeHost.AdditionalProps["username"],
					pipeHost.AdditionalProps["password"],
					hostAddress,   // Change localhost to host.docker.internal
					pipeHost.AdditionalProps["messaging_port"],
					pipeName,   // add the pipe name
					pipeType,   // add the pipe type (queue or topic)
				)

				// Set the environment variable for the pipe name (e.g., "in": "reutlingenPipe")
				envVars = append(envVars, fmt.Sprintf("%s=%s", parts[0], value))

				// Set the routingKey environment variable if defined
				if routingKey != "" {
					envVars = append(envVars, fmt.Sprintf("%sRoutingKey=%s", parts[0], routingKey))
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
				// Handle file-based config by finding the absolute path
				filePath := filepath.Join(baseDir, value)
				absoluteFilePath, err := filepath.Abs(filePath)
				if err != nil {
					fmt.Printf("failed to get absolute path for %s: %v\n", filePath, err)
					continue
				}

				volumeName := strings.ToLower(filter.Name + "-" + config.Name)
				volumes = append(volumes, volumeName)
				volumeMounts = append(volumeMounts, fmt.Sprintf("%s:/etc/config/criteria", absoluteFilePath))

				// Update the value to point to the new mount path
				value = "/etc/config/criteria"
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
