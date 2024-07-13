package models

type Model struct {
	Pipes              Pipes               `yaml:"pipes"`
	Filters            []Filter            `yaml:"filters"`
	Hosts              Hosts               `yaml:"hosts"`
	FilterTypes        []FilterType        `yaml:"filterTypes"`
	DeploymentArtifacts []DeploymentArtifact `yaml:"deploymentArtifacts"`
}

type Pipes struct {
	Queues []Queue `yaml:"queues"`
}

type Queue struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Protocol string `yaml:"protocol"`
	Configs  string `yaml:"configs,omitempty"`
}

type Filter struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Host     string   `yaml:"host"`
	Type     string   `yaml:"type"`
	Mappings []string `yaml:"mappings"`
	Artifact string   `yaml:"artifact,omitempty"`
}

type Hosts struct {
	PipeHosts   []Host `yaml:"pipeHosts"`
	FilterHosts []Host `yaml:"filterHosts"`
}

type Host struct {
	ID     string `yaml:"id"`
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Configs string `yaml:"configs,omitempty"`
}

type FilterType struct {
	Name        string             `yaml:"name"`
	Artifact    string             `yaml:"artifact,omitempty"`
	DerivesFrom string             `yaml:"derivesFrom,omitempty"`
	Configs     *FilterTypeConfigs `yaml:"configs,omitempty"`
}

type FilterTypeConfigs struct {
	Enforces []string               `yaml:"enforces,omitempty"`
	Optional map[string]interface{} `yaml:"optional,omitempty"`
}

type DeploymentArtifact struct {
	Name          string   `yaml:"name"`
	Type          string   `yaml:"type"`
	Image         string   `yaml:"image"`
	InternalPipes []string `yaml:"internalPipes"`
}
