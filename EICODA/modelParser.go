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
	"eicoda/utils" // Ensure this import is included
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

	var rawHostTypes struct {
		Hosts models.HostTypes `yaml:"hosts"`
	}
	err = yaml.Unmarshal(data, &rawHostTypes)
	if err != nil {
		log.Fatalf("Error unmarshalling host types: %v", err)
	}

	parser.hostTypes = rawHostTypes.Hosts
	fmt.Println("Loaded host types.")
}

// Parse parses the YAML file at the given path and returns a merged Model
func (parser *ModelParser) Parse(path string) (*models.Model, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return parser.parseData(data)
}

// ParseFromString parses the YAML content from a string and returns a merged Model
func (parser *ModelParser) ParseFromString(content string) (*models.Model, error) {
	data := []byte(content)
	return parser.parseData(data)
}

// parseData parses the YAML data and returns a merged Model
func (parser *ModelParser) parseData(data []byte) (*models.Model, error) {
	var model models.Model
	err := yaml.Unmarshal(data, &model)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling YAML: %w", err)
	}

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

	parser.resolveInheritance(&combinedTypes)

	err = parser.mergeModels(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

	err = parser.performChecks(&model)
	if err != nil {
		return nil, fmt.Errorf("parsing model failed: %w", err)
	}

	err = parser.applyFilterTypeArtifacts(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

	err = parser.checkFilterMappings(&model, &combinedTypes)
	if err != nil {
		return nil, err
	}

	fmt.Println("Parsed and merged model successfully.")
	return &model, nil
}

// mergeModels merges the parsed model with the loaded types and artifacts
func (parser *ModelParser) mergeModels(model *models.Model, combinedTypes *models.CombinedTypes) error {
	// Merge FilterTypes
	filterTypeNames := make(map[string]bool)
	for _, ft := range model.FilterTypes {
		filterTypeNames[ft.Name] = true
	}
	for _, ft := range combinedTypes.FilterTypes {
		if ft.Name == "" {
			continue
		}
		if filterTypeNames[ft.Name] {
			return fmt.Errorf("duplicate filter type name found: %s", ft.Name)
		}
		model.FilterTypes = append(model.FilterTypes, ft)
	}

	// Merge DeploymentArtifacts
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

	// Merge Hosts
	err := parser.mergeHosts(model, &combinedTypes.Hosts)
	if err != nil {
		return err
	}

	return nil
}

// mergeHosts merges the hosts from the combined types with those in the model
func (parser *ModelParser) mergeHosts(model *models.Model, combinedHosts *models.Hosts) error {
	hostIDs := make(map[string]bool)
	for _, host := range model.Hosts.PipeHosts {
		hostIDs[host.ID] = true
	}
	for _, host := range combinedHosts.PipeHosts {
		if hostIDs[host.ID] {
			return fmt.Errorf("duplicate pipeHost ID found: %s", host.ID)
		}
		model.Hosts.PipeHosts = append(model.Hosts.PipeHosts, host)
	}

	for _, host := range model.Hosts.FilterHosts {
		hostIDs[host.ID] = true
	}
	for _, host := range combinedHosts.FilterHosts {
		if hostIDs[host.ID] {
			return fmt.Errorf("duplicate filterHost ID found: %s", host.ID)
		}
		model.Hosts.FilterHosts = append(model.Hosts.FilterHosts, host)
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
		parser.inheritFilterTypeProperties(filterTypeMap, &combinedTypes.FilterTypes[i])
	}
}

func (parser *ModelParser) inheritFilterTypeProperties(filterTypeMap map[string]*models.FilterType, ft *models.FilterType) {
	if ft.DerivedFrom == "" {
		return
	}

	parent, exists := filterTypeMap[ft.DerivedFrom]
	if !exists {
		fmt.Printf("Parent filter type %s not found for filter type %s\n", ft.DerivedFrom, ft.Name)
		return
	}

	parser.inheritFilterTypeProperties(filterTypeMap, parent)

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
			return fmt.Errorf("filter type %s not found for filter %s", filter.Type, filter.Name)
		}

		if filter.AdditionalProps == nil {
			filter.AdditionalProps = make(map[string]string)
		}

		for _, config := range filterType.Configs {
			_, exists := filter.AdditionalProps[config.Name]
			if !exists || filter.AdditionalProps[config.Name] == "" {
				if config.Default != nil {
					filter.AdditionalProps[config.Name] = fmt.Sprintf("%v", config.Default)
				} else {
					return fmt.Errorf("filter %s of type %s is missing required property: %s", filter.Name, filter.Type, config.Name)
				}
			}
		}

		for prop := range filter.AdditionalProps {
			allowed := false
			for _, config := range filterType.Configs {
				if config.Name == prop {
					allowed = true
					break
				}
			}
			if !allowed {
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
			continue
		}

		filterType, exists := filterTypeMap[filter.Type]
		if !exists {
			return fmt.Errorf("filter type %s not found for filter %s", filter.Type, filter.Name)
		}

		if filterType.Artifact != "" && filter.Artifact == "" {
			filter.Artifact = filterType.Artifact
		}

		model.Filters[i] = filter
	}

	return nil
}

// performChecks performs various correctness checks on the parsed model
func (parser *ModelParser) performChecks(model *models.Model) error {
	err := parser.checkForDuplicateIDs(model)
	if err != nil {
		return err
	}

	err = parser.checkQueueHosts(model)
	if err != nil {
		return err
	}

	err = parser.checkQueueProtocols(model)
	if err != nil {
		return err
	}

	err = parser.checkFilterHosts(model)
	if err != nil {
		return err
	}

	err = parser.checkHostTypes(model)
	if err != nil {
		return err
	}

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

	for _, topic := range model.Pipes.Topics {
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

	for _, topic := range model.Pipes.Topics {
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
	for _, topic := range model.Pipes.Topics {
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
	for _, topic := range model.Pipes.Topics {
		definedPipes[topic.Name] = true
	}

	artifactMap := make(map[string]models.DeploymentArtifact)
	for _, artifact := range combinedTypes.DeploymentArtifacts {
		artifactMap[artifact.Name] = artifact
	}
	for _, artifact := range model.DeploymentArtifacts {
		artifactMap[artifact.Name] = artifact
	}

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

			externalParts := strings.Split(parts[1], "->")
			pipeName := externalParts[0]
			if !definedPipes[pipeName] {
				return fmt.Errorf("mapping target pipe %s not defined in queues or topics", pipeName)
			}

			// Check if the pipe is a queue or topic
			queue := utils.FindQueueByName(model.Pipes.Queues, pipeName)
			topic := utils.FindTopicByName(model.Pipes.Topics, pipeName)

			if queue != nil && len(externalParts) > 1 {
				return fmt.Errorf("routingKey not allowed on queue %s", pipeName)
			}

			if topic != nil && len(externalParts) > 1 {
				// Valid routingKey provided for topic
				continue
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
		}

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
		valid := false
		for _, ht := range parser.hostTypes.PipeHosts {
			if host.Type == ht.Name {
				valid = true

				for prop := range host.AdditionalProps {
					if !contains(ht.Configs, prop) {
						return fmt.Errorf("pipeHost %s of type %s has invalid property: %s", host.Name, host.Type, prop)
					}
				}

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

	for _, host := range model.Hosts.FilterHosts {
		valid := false
		for _, ht := range parser.hostTypes.FilterHosts {
			if host.Type == ht.Name {
				valid = true

				for prop := range host.AdditionalProps {
					if !contains(ht.Configs, prop) {
						return fmt.Errorf("filterHost %s of type %s has invalid property: %s", host.Name, host.Type, prop)
					}
				}

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
