package main

import (
	"fmt"
	"io/ioutil"
	"log"
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

// Parse parses the YAML file at the given path and returns a Model
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

	fmt.Printf("Parsed Model: %+v\n", model)
	return &model, nil
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
