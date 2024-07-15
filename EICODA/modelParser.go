package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"eicoda/models"
)

// ModelParser handles parsing of deployment configuration files
type ModelParser struct {
	hostTypes models.HostTypes
}

// NewModelParser creates a new instance of ModelParser
func NewModelParser() *ModelParser {
	parser := &ModelParser{}
	parser.loadHostTypes()
	return parser
}

// loadHostTypes loads the host types from the hostTypes.yaml file
func (parser *ModelParser) loadHostTypes() {
	data, err := ioutil.ReadFile(filepath.Join("repositoryControllers", "hostTypes.yaml"))
	if err != nil {
		log.Fatalf("Error reading host types file: %v", err)
	}

	log.Printf("Raw hostTypes.yaml content: %s\n", string(data))

	var rawHostTypes struct {
		Hosts models.HostTypes `yaml:"hosts"`
	}
	err = yaml.Unmarshal(data, &rawHostTypes)
	if err != nil {
		log.Fatalf("Error unmarshalling host types: %v", err)
	}

	parser.hostTypes = rawHostTypes.Hosts
	log.Printf("Loaded Host Types: %+v\n", parser.hostTypes)
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

	log.Printf("Parsed Model: %+v\n", model)

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

	log.Printf("Raw types.yaml/mergedTypes.yaml content: %s\n", string(typesData))

	var combinedTypes models.CombinedTypes
	err = yaml.Unmarshal(typesData, &combinedTypes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse types file: %w", err)
	}

	log.Printf("CombinedTypes before merging: %+v\n", combinedTypes)

	// Resolve inheritance for filter types
	parser.resolveInheritance(&combinedTypes)

	// Merge the parsed model with the loaded types and artifacts
	err = parser.mergeModels(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

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

// resolveInheritance resolves the inheritance hierarchy for filter types
func (parser *ModelParser) resolveInheritance(combinedTypes *models.CombinedTypes) {
	filterTypeMap := make(map[string]*models.FilterType)
	for i, ft := range combinedTypes.FilterTypes {
		filterTypeMap[ft.Name] = &combinedTypes.FilterTypes[i]
	}

	for i := range combinedTypes.FilterTypes {
		log.Printf("Resolving inheritance for filter type: %s", combinedTypes.FilterTypes[i].Name)
		parser.inheritFilterTypeProperties(filterTypeMap, &combinedTypes.FilterTypes[i])
	}
}

func (parser *ModelParser) inheritFilterTypeProperties(filterTypeMap map[string]*models.FilterType, ft *models.FilterType) {
	if ft.DerivedFrom == "" {
		return
	}

	parent, exists := filterTypeMap[ft.DerivedFrom]
	if !exists {
		log.Printf("Parent filter type %s not found for filter type %s", ft.DerivedFrom, ft.Name)
		return
	}

	log.Printf("Inheriting properties from parent filter type %s to %s", parent.Name, ft.Name)

	// Recursively inherit from the parent first
	parser.inheritFilterTypeProperties(filterTypeMap, parent)

	// Inherit configurations
	if parent.Configs != nil {
		for _, parentConfig := range parent.Configs {
			exists := false
			for _, config := range ft.Configs {
				if config.Name == parentConfig.Name {
					exists = true
					break
				}
			}
			if !exists {
				ft.Configs = append(ft.Configs, parentConfig)
			}
		}
	}

	// Inherit artifact if not set
	if ft.Artifact == "" {
		ft.Artifact = parent.Artifact
	}
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
			log.Printf("Filter type %s not found for filter %s", filter.Type, filter.Name)
			return fmt.Errorf("filter type %s not found for filter %s", filter.Type, filter.Name)
		}

		// Initialize AdditionalProps if it is nil
		if filter.AdditionalProps == nil {
			filter.AdditionalProps = make(map[string]string)
		}

		for _, config := range filterType.Configs {
			_, exists := filter.AdditionalProps[config.Name]
			if !exists || filter.AdditionalProps[config.Name] == "" {
				if config.Default != nil {
					log.Printf("Setting default value for %s: %v", config.Name, config.Default)
					filter.AdditionalProps[config.Name] = fmt.Sprintf("%v", config.Default)
				} else {
					log.Printf("Filter %s of type %s is missing required property: %s", filter.Name, filter.Type, config.Name)
					return fmt.Errorf("filter %s of type %s is missing required property: %s", filter.Name, filter.Type, config.Name)
				}
			}
		}

		// Check for any additional properties not allowed by the config
		for prop := range filter.AdditionalProps {
			allowed := false
			for _, config := range filterType.Configs {
				if config.Name == prop {
					allowed = true
					break
				}
			}
			if !allowed {
				log.Printf("Filter %s of type %s has an invalid property: %s", filter.Name, filter.Type, prop)
				return fmt.Errorf("filter %s of type %s has an invalid property: %s", filter.Name, filter.Type, prop)
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
			log.Printf("Filter type %s not found for filter %s", filter.Type, filter.Name)
			return fmt.Errorf("filter type %s not found for filter %s", filter.Type, filter.Name)
		}

		// Apply inherited artifact if filter's artifact is not explicitly set
		if filterType.Artifact != "" && filter.Artifact == "" {
			filter.Artifact = filterType.Artifact
			log.Printf("Filter %s of type %s is using inherited artifact %s\n", filter.Name, filter.Type, filterType.Artifact)
		} else if filter.Artifact != "" {
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

	// Check filter type enforcements
	err = parser.checkFilterTypeEnforcements(model)
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

	for _, topic := range model.Pipes.Topics { // Add this block to check for duplicate topic IDs
		if idSet[topic.ID] {
			return fmt.Errorf("duplicate ID found: %s", topic.ID)
		}
		idSet[topic.ID] = true
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

	for _, topic := range model.Pipes.Topics { // Add this block to check if the host field of each topic refers to a defined name of a pipeHost
		if !pipeHosts[topic.Host] {
			return fmt.Errorf("topic host %s is not defined as a pipeHost", topic.Host)
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
	for _, topic := range model.Pipes.Topics { // Add this block to check if the protocol is either amqp or mqtt for topics
		if topic.Protocol != "amqp" && topic.Protocol != "mqtt" {
			return fmt.Errorf("topic protocol %s is invalid for topic %s", topic.Protocol, topic.Name)
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
	for _, topic := range model.Pipes.Topics { // Add this block to include topics in the defined pipes
		definedPipes[topic.Name] = true
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

		allowedProtocols := strings.Split(artifact.Protocol, ",")
		protocolMap := make(map[string]bool)
		for _, p := range allowedProtocols {
			protocolMap[p] = true
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

			pipeName := parts[1]
			if !definedPipes[pipeName] {
				return fmt.Errorf("mapping target pipe %s not defined in queues or topics", pipeName)
			}

			var pipeProtocol string
			for _, queue := range model.Pipes.Queues {
				if queue.Name == pipeName {
					pipeProtocol = queue.Protocol
					break
				}
			}
			for _, topic := range model.Pipes.Topics { // Add this block to get the protocol for topics
				if topic.Name == pipeName {
					pipeProtocol = topic.Protocol
					break
				}
			}

			if !protocolMap[pipeProtocol] {
				return fmt.Errorf("protocol %s of pipe %s is not allowed by deployment artifact %s", pipeProtocol, pipeName, artifact.Name)
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
	log.Printf("Validating pipeHosts: %+v\n", model.Hosts.PipeHosts)
	for _, host := range model.Hosts.PipeHosts {
		valid := false
		for _, ht := range parser.hostTypes.PipeHosts {
			log.Printf("Checking host type: %s against %s\n", host.Type, ht.Name)
			if host.Type == ht.Name {
				valid = true

				// Check for additional properties not defined in the host type configs
				for prop := range host.AdditionalProps {
					if !contains(ht.Configs, prop) {
						return fmt.Errorf("pipeHost %s of type %s has invalid property: %s", host.Name, host.Type, prop)
					}
				}

				// Check for missing required properties
				for _, config := range ht.Configs {
					if _, exists := host.AdditionalProps[config]; !exists {
						return fmt.Errorf("pipeHost %s of type %s is missing required property: %s", host.Name, host.Type, config)
					}
				}
				break
			}
		}
		if !valid {
			return fmt.Errorf("pipeHost %s has invalid type %s", host.Name, host.Type)
		}
	}

	log.Printf("Validating filterHosts: %+v\n", model.Hosts.FilterHosts)
	for _, host := range model.Hosts.FilterHosts {
		valid := false
		for _, ht := range parser.hostTypes.FilterHosts {
			log.Printf("Checking host type: %s against %s\n", host.Type, ht.Name)
			if host.Type == ht.Name {
				valid = true

				// Check for additional properties not defined in the host type configs
				for prop := range host.AdditionalProps {
					if !contains(ht.Configs, prop) {
						return fmt.Errorf("filterHost %s of type %s has invalid property: %s", host.Name, host.Type, prop)
					}
				}

				// Check for missing required properties
				for _, config := range ht.Configs {
					if _, exists := host.AdditionalProps[config]; !exists {
						return fmt.Errorf("filterHost %s of type %s is missing required property: %s", host.Name, host.Type, config)
					}
				}
				break
			}
		}
		if !valid {
			return fmt.Errorf("filterHost %s has invalid type %s", host.Name, host.Type)
		}
	}

	return nil
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
