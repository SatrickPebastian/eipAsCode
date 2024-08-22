package repositoryControllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
	"eicoda/models"
)

//represents the structure of the combined types and artifacts YAML file
type CombinedTypes struct {
	FilterTypes         []models.FilterType         `yaml:"filterTypes"`
	DeploymentArtifacts []models.DeploymentArtifact `yaml:"deploymentArtifacts"`
	Hosts               models.Hosts                `yaml:"hosts"`
}

// handles operations related to filter types and artifacts
type TypeController struct {
	filterTypes         []models.FilterType
	deploymentArtifacts []models.DeploymentArtifact
	hosts               models.Hosts
}

func NewTypeController() *TypeController {
	tc := &TypeController{}
	tc.loadInitialData()
	return tc
}

//loads the types.yaml file
func (tc *TypeController) loadInitialData() {
	typesPath := filepath.Join("repositoryControllers", "types.yaml")
	typesData, err := ioutil.ReadFile(typesPath)
	if err != nil {
		fmt.Printf("failed to read types.yaml: %v\n", err)
		return
	}
	var combinedTypes CombinedTypes
	err = yaml.Unmarshal(typesData, &combinedTypes)
	if err != nil {
		fmt.Printf("failed to parse types.yaml: %v\n", err)
		return
	}
	tc.filterTypes = combinedTypes.FilterTypes
	tc.deploymentArtifacts = combinedTypes.DeploymentArtifacts
	tc.hosts = combinedTypes.Hosts
}

//reads the new types YAML file, validates the filter types and artifacts, and merges them with the existing types
func (tc *TypeController) AddType(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var newCombinedTypes CombinedTypes

	err = yaml.Unmarshal(data, &newCombinedTypes)
	if err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	for _, filterType := range newCombinedTypes.FilterTypes {
		if err := validateFilterType(filterType); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		for _, existingType := range tc.filterTypes {
			if existingType.Name == filterType.Name {
				return fmt.Errorf("duplicate filter type name found: %s", filterType.Name)
			}
		}

		tc.filterTypes = append(tc.filterTypes, filterType)
	}

	for _, artifact := range newCombinedTypes.DeploymentArtifacts {
		for _, existingArtifact := range tc.deploymentArtifacts {
			if existingArtifact.Name == artifact.Name {
				return fmt.Errorf("duplicate deployment artifact name found: %s", artifact.Name)
			}
		}

		tc.deploymentArtifacts = append(tc.deploymentArtifacts, artifact)
	}

	tc.mergeHosts(newCombinedTypes.Hosts)

	for _, filterType := range tc.filterTypes {
		if filterType.Artifact != "" {
			if !tc.isValidArtifact(filterType.Artifact) {
				return fmt.Errorf("invalid artifact specified: %s", filterType.Artifact)
			}
		}
	}

	err = tc.saveMergedTypes()
	if err != nil {
		return fmt.Errorf("failed to save merged types: %w", err)
	}

	combinedTypes := CombinedTypes{
		FilterTypes:         tc.filterTypes,
		DeploymentArtifacts: tc.deploymentArtifacts,
		Hosts:               tc.hosts,
	}
	combinedTypesJSON, err := json.MarshalIndent(combinedTypes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to convert combined types to JSON: %w", err)
	}

	fmt.Println("Successfully validated and merged filter types, deployment artifacts, and hosts:", string(combinedTypesJSON))

	return nil
}

//merges the new hosts with the existing ones, avoiding duplicates
func (tc *TypeController) mergeHosts(newHosts models.Hosts) {
	for _, newHost := range newHosts.PipeHosts {
		if !tc.hostExists(newHost, tc.hosts.PipeHosts) {
			tc.hosts.PipeHosts = append(tc.hosts.PipeHosts, newHost)
		}
	}

	for _, newHost := range newHosts.FilterHosts {
		if !tc.hostExists(newHost, tc.hosts.FilterHosts) {
			tc.hosts.FilterHosts = append(tc.hosts.FilterHosts, newHost)
		}
	}
}

// checks if a host already exists in the given slice of hosts
func (tc *TypeController) hostExists(newHost models.Host, existingHosts []models.Host) bool {
	for _, host := range existingHosts {
		if host.ID == newHost.ID {
			return true
		}
	}
	return false
}

//validates the structure of a filter type
func validateFilterType(ft models.FilterType) error {
	if ft.Name == "" {
		return fmt.Errorf("name is required")
	}

	if ft.Configs != nil {
		for _, config := range ft.Configs {
			if config.Name == "" {
				return fmt.Errorf("config name cannot be empty")
			}
		}
	}

	return nil
}

//checks if the artifact is valid
func (tc *TypeController) isValidArtifact(artifact string) bool {
	for _, da := range tc.deploymentArtifacts {
		if da.Name == artifact {
			return true
		}
	}
	return false
}

// saves the merged filter types, deployment artifacts, and hosts to mergedTypes.yaml
func (tc *TypeController) saveMergedTypes() error {
	mergedTypesFile := CombinedTypes{
		FilterTypes:         tc.filterTypes,
		DeploymentArtifacts: tc.deploymentArtifacts,
		Hosts:               tc.hosts,
	}

	data, err := yaml.Marshal(mergedTypesFile)
	if err != nil {
		return fmt.Errorf("failed to marshal merged types: %w", err)
	}

	mergedTypesPath := filepath.Join("repositoryControllers", "mergedTypes.yaml")
	err = ioutil.WriteFile(mergedTypesPath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write merged types file: %w", err)
	}

	return nil
}
