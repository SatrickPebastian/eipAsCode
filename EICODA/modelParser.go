package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"eicoda/models" // Adjust the import path according to your module name
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

	// Perform correctness checks
	err = parser.performChecks(&model)
	if err != nil {
		return nil, fmt.Errorf("Parsing model failed: %w", err)
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

	// Merge the parsed model with the loaded types and artifacts
	err = parser.mergeModels(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

	// Perform filter type enforcement checks
	err = parser.checkFilterTypeEnforcements(&model)
	if err != nil {
		return nil, err
	}

	// Check if filter types are valid
	err = parser.checkFilterTypes(&model)
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
		if filterType, exists := filterTypeMap[filter.Type]; exists {
			if filter.Type == "Custom" {
				continue // Allow any properties for Custom type
			}

			if filterType.Configs != nil {
				for _, enforced := range filterType.Configs.Enforces {
					if !parser.hasProperty(filter, enforced) {
						return fmt.Errorf("filter %s of type %s is missing required property: %s", filter.Name, filter.Type, enforced)
					}
				}

				// Check for invalid properties
				allowedProps := append(filterType.Configs.Enforces, "id", "name", "host", "type", "mappings", "artifact")
				for opt := range filterType.Configs.Optional {
					allowedProps = append(allowedProps, opt)
				}

				for prop := range filter.AdditionalProps {
					if !parser.isAllowedProperty(prop, allowedProps) {
						return fmt.Errorf("filter %s of type %s has an invalid property: %s", filter.Name, filter.Type, prop)
					}
				}
			}
		}
	}

	return nil
}

// checkFilterTypes checks if filter types are valid
func (parser *ModelParser) checkFilterTypes(model *models.Model) error {
	validTypes := make(map[string]bool)
	for _, ft := range model.FilterTypes {
		validTypes[ft.Name] = true
	}

	for _, filter := range model.Filters {
		if !validTypes[filter.Type] {
			return fmt.Errorf("filter %s has an invalid type: %s", filter.Name, filter.Type)
		}
	}

	return nil
}

// hasProperty checks if a filter has a specified property
func (parser *ModelParser) hasProperty(filter models.Filter, property string) bool {
	_, exists := filter.AdditionalProps[property]
	return exists
}

// isAllowedProperty checks if a property is allowed
func (parser *ModelParser) isAllowedProperty(property string, allowedProps []string) bool {
	for _, p := range allowedProps {
		if p == property {
			return true
		}
	}
	return false
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

	// Check filter mappings against deployment artifacts
	err = parser.checkFilterMappings(model)
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
func (parser *ModelParser) checkFilterMappings(model *models.Model) error {
	definedPipes := make(map[string]bool)
	for _, queue := range model.Pipes.Queues {
		definedPipes[queue.Name] = true
	}

	for _, filter := range model.Filters {
		for _, mapping := range filter.Mappings {
			parts := strings.Split(mapping, ":")
			if len(parts) != 2 {
				return fmt.Errorf("invalid mapping format: %s", mapping)
			}
			internalPipeFound := false
			for _, artifact := range model.DeploymentArtifacts {
				for _, internalPipe := range artifact.InternalPipes {
					if parts[0] == internalPipe {
						internalPipeFound = true
						break
					}
				}
				if internalPipeFound {
					break
				}
			}
			if !internalPipeFound {
				return fmt.Errorf("mapping internal pipe %s not defined in any deployment artifact", parts[0])
			}
			if !definedPipes[parts[1]] {
				return fmt.Errorf("mapping target pipe %s not defined in queues", parts[1])
			}
		}
	}

	// Ensure all internal pipes are covered in the mappings
	for _, artifact := range model.DeploymentArtifacts {
		internalPipesSet := make(map[string]bool)
		for _, internalPipe := range artifact.InternalPipes {
			internalPipesSet[internalPipe] = true
		}
		for _, filter := range model.Filters {
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
