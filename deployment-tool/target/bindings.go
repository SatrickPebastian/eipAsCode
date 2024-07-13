package target

import (
    "fmt"
    "os"
    "diploy/parser"
    "gopkg.in/yaml.v2"
)

func CreateBindings(config parser.Config) error {
    // Check if the components directory exists, create it if not
    componentsDir := "./target/components"
    if _, err := os.Stat(componentsDir); os.IsNotExist(err) {
        // Create directory if it does not exist
        if err := os.MkdirAll(componentsDir, 0755); err != nil {
            return fmt.Errorf("failed to create components directory: %v", err)
        }
    }

    for _, pipe := range config.Pipes {
        // Build the AMQP URL, include the port if specified
        var amqpURL string
        if amqpPort := config.DeploymentEnvironments[0].Pipes.AmqpPort; amqpPort != nil {
            amqpURL = fmt.Sprintf("amqp://%s:%d", config.DeploymentEnvironments[0].Pipes.Address, *amqpPort)
        } else {
            amqpURL = fmt.Sprintf("amqp://%s:5672", config.DeploymentEnvironments[0].Pipes.Address)
        }

        component := struct {
            APIVersion string `yaml:"apiVersion"`
            Kind       string `yaml:"kind"`
            Metadata   struct {
                Name string `yaml:"name"`
            } `yaml:"metadata"`
            Spec struct {
                Type     string `yaml:"type"`
                Version  string `yaml:"version"`
                Metadata []struct {
                    Name  string `yaml:"name"`
                    Value string `yaml:"value"`
                } `yaml:"metadata"`
            } `yaml:"spec"`
        }{
            APIVersion: "dapr.io/v1alpha1",
            Kind:       "Component",
            Metadata: struct {
                Name string `yaml:"name"`
            }{Name: pipe.Name + "-binding"},
            Spec: struct {
                Type     string `yaml:"type"`
                Version  string `yaml:"version"`
                Metadata []struct {
                    Name  string `yaml:"name"`
                    Value string `yaml:"value"`
                } `yaml:"metadata"`
            }{
                Type:    "bindings.rabbitmq",
                Version: "v1",
                Metadata: []struct {
                    Name  string `yaml:"name"`
                    Value string `yaml:"value"`
                }{
                    {Name: "queueName", Value: pipe.Name},
                    {Name: "host", Value: amqpURL}, // Use the correctly formatted AMQP URL
                    {Name: "username", Value: config.DeploymentEnvironments[0].Pipes.Username},
                    {Name: "password", Value: config.DeploymentEnvironments[0].Pipes.Password},
                },
            },
        }

        data, err := yaml.Marshal(&component)
        if err != nil {
            return err
        }

        fileName := fmt.Sprintf("%s/%s.yaml", componentsDir, pipe.Name)
        if err := os.WriteFile(fileName, data, 0644); err != nil {
            return fmt.Errorf("failed to write file %s: %v", fileName, err)
        }
    }
    return nil
}
