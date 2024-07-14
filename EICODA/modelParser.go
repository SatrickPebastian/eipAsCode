package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
	"eicoda/models"
)

// ModelParser handles parsing of deployment configuration files
type ModelParser struct {
}

// NewModelParser creates a new instance of ModelParser
func NewModelParser() *ModelParser {
	return &ModelParser{}
}

// Parse parses the YAML file at the given path and returns a merged Model
func (parser *ModelParser) Parse(path string) (*models.Model, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading file: %v\n", err)
		return nil, err
	}

	var model models.Model
	err = yaml.Unmarshal(data, &model)
	if err != nil {
		log.Printf("Error unmarshalling YAML: %v\n", err)
		return nil, err
	}

	// Load types.yaml or mergedTypes.yaml
	typesFilePath := filepath.Join("repositoryControllers", "types.yaml")
	mergedTypesFilePath := filepath.Join("repositoryControllers", "mergedTypes.yaml")
	var typesData []byte

	if _, err := os.Stat(mergedTypesFilePath); err == nil {
		typesData, err = ioutil.ReadFile(mergedTypesFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read mergedTypes.yaml: %w", err)
		}
	} else {
		typesData, err = ioutil.ReadFile(typesFilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read types.yaml: %w", err)
		}
	}

	var combinedTypes models.CombinedTypes
	err = yaml.Unmarshal(typesData, &combinedTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse types file: %w", err)
	}

	// Add log to see combinedTypes content
	log.Printf("CombinedTypes before merging: %+v\n", combinedTypes)

	// Merge the parsed model with the loaded types and artifacts
	err = parser.mergeModels(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

	// Add log to see combinedTypes content after merging
	log.Printf("CombinedTypes after merging: %+v\n", combinedTypes)

	// Perform correctness checks
	err = parser.performChecks(&model)
	if err != nil {
		return nil, fmt.Errorf("Parsing model failed: %w", err)
	}

	// Apply artifacts and mappings from filter types if not set in the filter
	err = parser.applyFilterTypeArtifacts(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

	// Perform additional checks on mappings
	err = parser.checkFilterMappings(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Parsed and Merged Model: %+v\n", model)
	return &model, nil
}

// mergeModels merges the parsed model with the loaded types and artifacts
func (parser *ModelParser) mergeModels(model *models.Model, combinedTypes *models.CombinedTypes) error {
	// Check for duplicate filter type names
	filterTypeNames := make(map[string]bool)
	for _, ft := range model.FilterTypes {
		filterTypeNames[ft.Name] = true
	}
	for _, ft := range combinedTypes.FilterTypes {
		if ft.Name == "" {
			continue // Ignore empty filter types
		}
		if filterTypeNames[ft.Name] {
			return fmt.Errorf("duplicate filter type name found: %s", ft.Name)
		}
		model.FilterTypes = append(model.FilterTypes, ft)
	}

	// Check for duplicate deployment artifact names
	deploymentArtifactNames := make(map[string]bool)
	for _, da := range model.DeploymentArtifacts {
		deploymentArtifactNames[da.Name] = true
	}
	for _, da := range combinedTypes.DeploymentArtifacts {
		if deploymentArtifactNames[da.Name] {
			return fmt.Errorf("duplicate deployment artifact name found: %s", da.Name)
		}
		model.DeploymentArtifacts = append(model.DeploymentArtifacts, da)
	}

	return nil
}

// checkFilterTypeEnforcements checks if filters have the required properties based on their type
func (parser *ModelParser) checkFilterTypeEnforcements(model *models.Model) error {
	filterTypeMap := make(map[string]models.FilterType)
	for _, ft := range model.FilterTypes {
		filterTypeMap[ft.Name] = ft
	}

	for _, filter := range model.Filters {
		if filter.Type == "Custom" {
			continue
		}

		filterType, exists := filterTypeMap[filter.Type]
		if !exists {
			return fmt.Errorf("filter type %s not found for filter %s", filter.Type, filter.Name)
		}

		filterValue := reflect.ValueOf(filter)
		for _, enforced := range filterType.Configs.Enforces {
			if !filterValue.FieldByName(enforced).IsValid() {
				return fmt.Errorf("filter %s of type %s is missing required property: %s", filter.Name, filter.Type, enforced)
			}
		}
	}

	return nil
}

// applyFilterTypeArtifacts applies artifacts and mappings from filter types if not set in the filter
func (parser *ModelParser) applyFilterTypeArtifacts(model *models.Model, combinedTypes *models.CombinedTypes) error {
	filterTypeMap := make(map[string]models.FilterType)
	for _, ft := range combinedTypes.FilterTypes {
		filterTypeMap[ft.Name] = ft
	}

	for i, filter := range model.Filters {
		if filter.Type == "Custom" && filter.Artifact != "" {
			log.Printf("Custom filter %s is using artifact %s\n", filter.Name, filter.Artifact)
			continue
		}

		filterType, exists := filterTypeMap[filter.Type]
		if !exists {
			return fmt.Errorf("filter type %s not found for filter %s", filter.Type, filter.Name)
		}

		if filter.Artifact == "" && filterType.Artifact != "" {
			filter.Artifact = filterType.Artifact
			log.Printf("Filter %s of type %s is using artifact %s from filter type\n", filter.Name, filter.Type, filterType.Artifact)
		}

		if filter.Artifact != "" {
			log.Printf("Filter %s is using artifact %s\n", filter.Name, filter.Artifact)
		}

		model.Filters[i] = filter
	}

	return nil
}

// performChecks performs various correctness checks on the parsed model
func (parser *ModelParser) performChecks(model *models.Model) error {
	// Check for duplicate IDs
	err := parser.checkForDuplicateIDs(model)
	if err != nil {
		return err
	}

	// Check if the host field of each queue refers to a defined name of a pipeHost
	err = parser.checkQueueHosts(model)
	if err != nil {
		return err
	}

	// Check if the protocol is either amqp or mqtt
	err = parser.checkQueueProtocols(model)
	if err != nil {
		return err
	}

	// Check if the host field of each filter refers to a defined name of a filterHost
	err = parser.checkFilterHosts(model)
	if err != nil {
		return err
	}

	// Check the types of pipeHosts and filterHosts
	err = parser.checkHostTypes(model)
	if err != nil {
		return err
	}

	return nil
}

// checkForDuplicateIDs checks for duplicate IDs across all objects
func (parser *ModelParser) checkForDuplicateIDs(model *models.Model) error {
	idSet := make(map[string]bool)

	for _, queue := range model.Pipes.Queues {
		if idSet[queue.ID] {
			return fmt.Errorf("duplicate ID found: %s", queue.ID)
		}
		idSet[queue.ID] = true
	}

	for _, filter := range model.Filters {
		if idSet[filter.ID] {
			return fmt.Errorf("duplicate ID found: %s", filter.ID)
		}
		idSet[filter.ID] = true
	}

	for _, host := range model.Hosts.PipeHosts {
		if idSet[host.ID] {
			return fmt.Errorf("duplicate ID found: %s", host.ID)
		}
		idSet[host.ID] = true
	}

	for _, host := range model.Hosts.FilterHosts {
		if idSet[host.ID] {
			return fmt.Errorf("duplicate ID found: %s", host.ID)
		}
		idSet[host.ID] = true
	}

	return nil
}

// checkQueueHosts checks if the host field of each queue refers to a defined name of a pipeHost
func (parser *ModelParser) checkQueueHosts(model *models.Model) error {
	pipeHosts := make(map[string]bool)
	for _, host := range model.Hosts.PipeHosts {
		pipeHosts[host.Name] = true
	}

	for _, queue := range model.Pipes.Queues {
		if !pipeHosts[queue.Host] {
			return fmt.Errorf("queue host %s is not defined as a pipeHost", queue.Host)
		}
	}

	return nil
}

// checkQueueProtocols checks if the protocol is either amqp or mqtt
func (parser *ModelParser) checkQueueProtocols(model *models.Model) error {
	for _, queue := range model.Pipes.Queues {
		if queue.Protocol != "amqp" && queue.Protocol != "mqtt" {
			return fmt.Errorf("queue protocol %s is invalid for queue %s", queue.Protocol, queue.Name)
		}
	}
	return nil
}

// checkFilterHosts checks if the host field of each filter refers to a defined name of a filterHost
func (parser *ModelParser) checkFilterHosts(model *models.Model) error {
	filterHosts := make(map[string]bool)
	for _, host := range model.Hosts.FilterHosts {
		filterHosts[host.Name] = true
	}

	for _, filter := range model.Filters {
		if !filterHosts[filter.Host] {
			return fmt.Errorf("filter host %s is not defined as a filterHost", filter.Host)
		}
	}

	return nil
}

// checkFilterMappings checks if filter mappings are correct based on deployment artifacts
func (parser *ModelParser) checkFilterMappings(model *models.Model, combinedTypes *models.CombinedTypes) error {
	definedPipes := make(map[string]bool)
	for _, queue := range model.Pipes.Queues {
		definedPipes[queue.Name] = true
	}

	artifactMap := make(map[string]models.DeploymentArtifact)
	for _, artifact := range combinedTypes.DeploymentArtifacts {
		artifactMap[artifact.Name] = artifact
	}
	for _, artifact := range model.DeploymentArtifacts {
		artifactMap[artifact.Name] = artifact
	}

	log.Printf("ArtifactMap: %+v\n", artifactMap)

	for _, filter := range model.Filters {
		artifact, exists := artifactMap[filter.Artifact]
		if !exists {
			return fmt.Errorf("artifact %s not found for filter %s", filter.Artifact, filter.Name)
		}

		for _, mapping := range filter.Mappings {
			parts := strings.Split(mapping, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid mapping format: %s", mapping)
			}
			internalPipeFound := false
			for _, internalPipe := range artifact.InternalPipes {
				if parts[0] == internalPipe {
					internalPipeFound = true
					break
				}
			}
			if !internalPipeFound {
				return fmt.Errorf("mapping internal pipe %s not defined in deployment artifact %s", parts[0], artifact.Name)
			}
			if !definedPipes[parts[1]] {
				return fmt.Errorf("mapping target pipe %s not defined in queues", parts[1])
			}
		}

		// Ensure all internal pipes are covered in the mappings
		internalPipesSet := make(map[string]bool)
		for _, internalPipe := range artifact.InternalPipes {
			internalPipesSet[internalPipe] = true
		}
		mappedPipes := make(map[string]bool)
		for _, mapping := range filter.Mappings {
			parts := strings.Split(mapping, ":")
			mappedPipes[parts[0]] = true
		}
		for internalPipe := range internalPipesSet {
			if !mappedPipes[internalPipe] {
				return fmt.Errorf("internal pipe %s is missing in filter mappings for deployment artifact %s", internalPipe, artifact.Name)
			}
		}
	}

	return nil
}

// checkHostTypes checks if pipeHosts have type RabbitMQ and filterHosts have type Kubernetes or DockerCompose
func (parser *ModelParser) checkHostTypes(model *models.Model) error {
	for _, host := range model.Hosts.PipeHosts {
		if host.Type != "RabbitMQ" {
			return fmt.Errorf("pipeHost %s has invalid type %s, expected RabbitMQ", host.Name, host.Type)
		}
	}

	for _, host := range model.Hosts.FilterHosts {
		if host.Type != "Kubernetes" && host.Type != "DockerCompose" {
			return fmt.Errorf("filterHost %s has invalid type %s, expected Kubernetes or DockerCompose", host.Name, host.Type)
		}
	}

	return nil
}
