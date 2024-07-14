package models

type Model struct {
	Pipes struct {
		Queues []Queue `yaml:"queues"`
	} `yaml:"pipes"`
	Filters            []Filter            `yaml:"filters"`
	Hosts              Hosts               `yaml:"hosts"`
	FilterTypes        []FilterType        `yaml:"filterTypes"`
	DeploymentArtifacts []DeploymentArtifact `yaml:"deploymentArtifacts"`
}

type Queue struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Protocol string `yaml:"protocol"`
	Configs  string `yaml:"configs"`
}

type Filter struct {
	ID                string            `yaml:"id"`
	Name              string            `yaml:"name"`
	Host              string            `yaml:"host"`
	Type              string            `yaml:"type"`
	Mappings          []string          `yaml:"mappings"`
	Artifact          string            `yaml:"artifact"`
	AdditionalProps   map[string]string `yaml:",inline"`
}

type Hosts struct {
	PipeHosts   []Host `yaml:"pipeHosts"`
	FilterHosts []Host `yaml:"filterHosts"`
}

type Host struct {
	ID      string `yaml:"id"`
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Configs string `yaml:"configs"`
}

type FilterType struct {
	Name        string             `yaml:"name"`
	Artifact    string             `yaml:"artifact,omitempty"`
	Configs     []FilterConfig     `yaml:"configs,omitempty"`
	DerivedFrom string             `yaml:"derivedFrom,omitempty"`
}

type FilterConfig struct {
	Name    string      `yaml:"name"`
	Default interface{} `yaml:"default,omitempty"`
}

type DeploymentArtifact struct {
	Name          string   `yaml:"name"`
	Type          string   `yaml:"type"`
	Image         string   `yaml:"image"`
	Protocol      string   `yaml:"protocol"`
	InternalPipes []string `yaml:"internalPipes"`
}

type CombinedTypes struct {
	FilterTypes        []FilterType        `yaml:"filterTypes"`
	DeploymentArtifacts []DeploymentArtifact `yaml:"deploymentArtifacts"`
}
