package models

type Model struct {
	Pipes struct {
		Queues []Queue `yaml:"queues"`
	} `yaml:"pipes"`
	Filters           []Filter           `yaml:"filters"`
	Hosts             Hosts              `yaml:"hosts"`
	FilterTypes       []FilterType       `yaml:"filterTypes"`
	DeploymentArtifacts DeploymentArtifact `yaml:"deploymentArtifacts"`
}

type Queue struct {
	ID       string `yaml:"id"`
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	Protocol string `yaml:"protocol"`
	Configs  string `yaml:"configs"`
}

type Filter struct {
	ID       string   `yaml:"id"`
	Name     string   `yaml:"name"`
	Host     string   `yaml:"host"`
	Type     string   `yaml:"type"`
	Mappings []string `yaml:"mappings"`
	Artifact string   `yaml:"artifact"`
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
	Name        string `yaml:"name"`
	Artifact    string `yaml:"artifact"`
	Configs     string `yaml:"configs"`
	DerivedFrom string `yaml:"derivedFrom"`
}

type DeploymentArtifact struct {
	Name          string   `yaml:"name"`
	Type          string   `yaml:"type"`
	Image         string   `yaml:"image"`
	InternalPipes []string `yaml:"internalPipes"`
}
