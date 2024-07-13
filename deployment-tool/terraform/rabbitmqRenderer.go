package terraform

import (
    "fmt"
    "os"
    "diploy/parser"
)

type RabbitMQRenderer struct{}

func (r RabbitMQRenderer) CreateTerraformFile(config parser.Config) error {
    file, err := os.Create("terraform.tf")
    if err != nil {
        return err
    }
    defer file.Close()

    // Determine the endpoint to use
    endpoint := config.DeploymentEnvironments[0].Pipes.Address
    if httpPort := config.DeploymentEnvironments[0].Pipes.HttpPort; httpPort != nil {
        // Use the HTTP port if specified for the management interface
        endpoint = fmt.Sprintf("http://%s:%d", endpoint, *httpPort)
    } else if amqpPort := config.DeploymentEnvironments[0].Pipes.AmqpPort; amqpPort != nil {
        // Use the AMQP port if specified for the standard AMQP connection
        endpoint = fmt.Sprintf("amqp://%s:%d", endpoint, *amqpPort)
    } else {
        // Default to standard AMQP port
        endpoint = fmt.Sprintf("amqp://%s:5672", endpoint)
    }

    fmt.Fprintf(file, `terraform {
required_providers {
  rabbitmq = {
    source  = "cyrilgdn/rabbitmq"
    version = "~> 1.8"
  }
}
}

provider "rabbitmq" {
endpoint = "%s"
username = "%s"
password = "%s"
}

`, endpoint, config.DeploymentEnvironments[0].Pipes.Username, config.DeploymentEnvironments[0].Pipes.Password)

    for _, pipe := range config.Pipes {
        durable := "false"
        if pipe.Persistent != nil && *pipe.Persistent {
            durable = "true"
        }

        fmt.Fprintf(file, `resource "rabbitmq_queue" "%s" {
name       = "%s"
vhost      = "/"
settings {
  durable   = %s
  auto_delete = false
}
}

`, pipe.Name, pipe.Name, durable)
    }

    return nil
}
