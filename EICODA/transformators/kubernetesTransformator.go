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

// Transform transforms the model to Kubernetes format
func (t *KubernetesTransformator) Transform(model *models.Model) error {
	var resources []interface{}

	for _, filter := range model.Filters {
		host := utils.FindHostByName(model.Hosts.FilterHosts, filter.Host)
		if host != nil && host.Type == "Kubernetes" {
			image := utils.FindArtifactImage(model.DeploymentArtifacts, filter.Artifact)
			deployment := createKubernetesDeployment(model, filter, image)
			resources = append(resources, deployment)
		}
	}

	// Generate the file at the project root
	outputPath := filepath.Join(".", "kubernetesModel.yaml")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes model file: %w", err)
	}
	defer file.Close()

	for i, resource := range resources {
		encoder := yaml.NewEncoder(file)
		err = encoder.Encode(resource)
		if err != nil {
			return fmt.Errorf("failed to write Kubernetes resource to file: %w", err)
		}
		encoder.Close()
		if i < len(resources)-1 {
			if _, err := file.WriteString("---\n"); err != nil {
				return fmt.Errorf("failed to write separator: %w", err)
			}
		}
	}

	fmt.Printf("Successfully created Kubernetes model file at %s\n", outputPath)
	return nil
}

func createKubernetesDeployment(model *models.Model, filter models.Filter, image string) map[string]interface{} {
	name := utils.SanitizeName(filter.Name)
	envVars := []map[string]string{}
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
						strings.ToLower(pipeHost.Type), // using the lowercase type name of the pipeHost
						pipeHost.AdditionalProps["messaging_port"],
						pipe.Name, // add the pipe name at the end
					)
					envVars = append(envVars, map[string]string{
						"name":  parts[0],
						"value": value,
					})
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
			envVars = append(envVars, map[string]string{
				"name":  config.Name,
				"value": utils.ConvertToProperType(value),
			})
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
							"name":  name,
							"image": image,
							"env":   envVars,
						},
					},
				},
			},
		},
	}

	return deployment
}
