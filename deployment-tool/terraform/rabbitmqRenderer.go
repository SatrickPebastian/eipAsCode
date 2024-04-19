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

`, config.DeploymentEnvironments[0].Pipes.Address, config.DeploymentEnvironments[0].Pipes.Username, config.DeploymentEnvironments[0].Pipes.Password)

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
