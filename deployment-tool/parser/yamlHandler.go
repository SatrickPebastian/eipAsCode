package parser

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
)

type Config struct {
    Pipes               []Pipe              `yaml:"pipes"`
    Filters             []Filter            `yaml:"filters"`
    DeploymentEnvironments []DeploymentEnvironment `yaml:"deployment-environments"`
}

type Pipe struct {
    Name        string `yaml:"name"`
    DLQ         *bool  `yaml:"dlq"`        
    Persistent  *bool  `yaml:"persistent"` 
}

type Filter struct {
    Name       string           `yaml:"name"`
    Type       string           `yaml:"type"`
    Properties FilterProperties `yaml:"properties"`
    Behavior   FilterBehavior   `yaml:"behavior"`
}

type FilterProperties struct {
    InputQueue   string   `yaml:"inputQueue"`
    OutputQueues []string `yaml:"outputQueues"`
}

type FilterBehavior struct {
    Conditions []Condition `yaml:"conditions"`
}

type Condition struct {
    Condition string `yaml:"condition"`
    Queue     string `yaml:"queue"`
}

type DeploymentEnvironment struct {
    Pipes   EnvironmentPipes   `yaml:"pipes"`
    Filters EnvironmentFilters `yaml:"filters"`
}

type EnvironmentPipes struct {
    Type       string  `yaml:"type"`
    Address    string  `yaml:"address"`
    HttpPort   *int    `yaml:"http-port"`
    AmqpPort   *int    `yaml:"amqp-port"`
    Username   string  `yaml:"username"`
    Password   string  `yaml:"password"`
}

type EnvironmentFilters struct {
    Type string `yaml:"type"`
}

func ReadYAMLConfig(filepath string) (Config, error) {
    var config Config
    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        return config, err
    }
    if err := yaml.Unmarshal(data, &config); err != nil {
        return config, err
    }
    return config, nil
}
