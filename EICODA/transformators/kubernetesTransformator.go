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

// KubernetesTransformator is responsible for transforming the model to Kubernetes format
type KubernetesTransformator struct{}

// Transform transforms the model to Kubernetes format and optionally writes to a file
func (t *KubernetesTransformator) Transform(model *models.Model, writeFile bool, baseDir string) (string, error) {
	var resources []interface{}

	for _, filter := range model.Filters {
		host := utils.FindHostByName(model.Hosts.FilterHosts, filter.Host)
		if host != nil && host.Type == "Kubernetes" {
			image := utils.FindArtifactImage(model.DeploymentArtifacts, filter.Artifact)
			deployment, configMap := createKubernetesDeployment(model, filter, image, baseDir)
			resources = append(resources, deployment)
			if configMap != nil {
				resources = append(resources, configMap)
			}
		}
	}

	var sb strings.Builder
	for i, resource := range resources {
		encoder := yaml.NewEncoder(&sb)
		err := encoder.Encode(resource)
		if err != nil {
			return "", fmt.Errorf("failed to encode Kubernetes resource: %w", err)
		}
		encoder.Close()
		if i < len(resources)-1 {
			sb.WriteString("---\n")
		}
	}

	// Write to file if writeFile is true
	if writeFile {
		outputPath := filepath.Join(".", "kubernetesModel.yaml")
		err := os.WriteFile(outputPath, []byte(sb.String()), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write Kubernetes model to file: %w", err)
		}
	}

	return sb.String(), nil
}

func createKubernetesDeployment(model *models.Model, filter models.Filter, image string, baseDir string) (map[string]interface{}, map[string]interface{}) {
	name := utils.SanitizeName(filter.Name)
	envVars := []map[string]interface{}{}
	volumeMounts := []map[string]interface{}{}
	volumes := []map[string]interface{}{}

	for _, mapping := range filter.Mappings {
		parts := strings.Split(mapping, ":")
		if len(parts) == 2 {
			pipeName := parts[1]
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
				value := fmt.Sprintf("%s://%s:%s@%s:%s,%s,%s",
					pipeProtocol,
					pipeHost.AdditionalProps["username"],
					pipeHost.AdditionalProps["password"],
					strings.ToLower(pipeHost.Type),
					pipeHost.AdditionalProps["messaging_port"],
					pipeName,   // add the pipe name
					pipeType,   // add the pipe type (queue or topic)
				)
				envVars = append(envVars, map[string]interface{}{
					"name":  parts[0],
					"value": value,
				})
			}
		}
	}

	// Add environment variables for filter type configs
	filterType := utils.FindFilterTypeByName(model.FilterTypes, filter.Type)
	var configMap map[string]interface{}
	if filterType != nil {
		for _, config := range filterType.Configs {
			value, exists := filter.AdditionalProps[config.Name]
			if !exists {
				value = fmt.Sprintf("%v", config.Default)
			}
			if config.File {
                // Handle file-based config
                filePath := filepath.Join(baseDir, value)
                fileContent, err := os.ReadFile(filePath)
                if err != nil {
                    fmt.Printf("failed to read file %s: %v", filePath, err)
                    continue
                }

				configMapName := strings.ToLower(name + "-" + config.Name)
				configMap = map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "ConfigMap",
					"metadata": map[string]interface{}{
						"name": configMapName,
					},
					"data": map[string]interface{}{
						config.Name: string(fileContent),
					},
				}

				volumeMounts = append(volumeMounts, map[string]interface{}{
					"name":      configMapName,
					"mountPath": fmt.Sprintf("/etc/config/%s", config.Name),
					"subPath":   config.Name,
				})

				volumes = append(volumes, map[string]interface{}{
					"name": configMapName,
					"configMap": map[string]interface{}{
						"name": configMapName,
					},
				})
			} else {
				envVars = append(envVars, map[string]interface{}{
					"name":  config.Name,
					"value": utils.ConvertToProperType(value),
				})
			}
		}
	}

	deployment := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": name,
		},
		"spec": map[string]interface{}{
			"replicas": 1,
			"selector": map[string]interface{}{
				"matchLabels": map[string]interface{}{
					"app": name,
				},
			},
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": name,
					},
				},
				"spec": map[string]interface{}{
					"containers": []map[string]interface{}{
						{
							"name":         name,
							"image":        image,
							"env":          envVars,
							"volumeMounts": volumeMounts,
						},
					},
					"volumes": volumes,
				},
			},
		},
	}

	return deployment, configMap
}
