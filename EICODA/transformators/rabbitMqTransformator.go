package transformators

import (
	"fmt"
	"os"
	"strings"

	"eicoda/models"
	"eicoda/utils"
)

type RabbitMqTransformator struct{}

func (t *RabbitMqTransformator) Transform(model *models.Model, writeFile bool, baseDir string) (string, error) {
	terraformResources := `
terraform {
  required_providers {
    rabbitmq = {
      source = "cyrilgdn/rabbitmq"
      version = "1.8.0"
    }
  }
}

provider "rabbitmq" {
  endpoint  = "%s"
  username  = "%s"
  password  = "%s"
}
`

	var endpoint, username, password string
	for _, host := range model.Hosts.PipeHosts {
		if host.Type == "RabbitMQ" {
			hostAddress := host.AdditionalProps["host_address"]
			managementPort := host.AdditionalProps["management_port"]
			endpoint = fmt.Sprintf("http://%s:%s", hostAddress, managementPort)
			username = host.AdditionalProps["username"]
			password = host.AdditionalProps["password"]
			break
		}
	}

	terraformResources = fmt.Sprintf(terraformResources, endpoint, username, password)

	for _, pipe := range model.Pipes.Queues {
		host := utils.FindHostByName(model.Hosts.PipeHosts, pipe.Host)
		if host != nil && host.Type == "RabbitMQ" {
			resource := createRabbitMqQueueResource(pipe, host)
			terraformResources += resource + "\n"
		}
	}

	for _, topic := range model.Pipes.Topics {
		host := utils.FindHostByName(model.Hosts.PipeHosts, topic.Host)
		if host != nil && host.Type == "RabbitMQ" {
			resource := createRabbitMqTopicResource(topic, host)
			terraformResources += resource + "\n"
		}
	}

	if writeFile {
		outputPath := "rabbitMqModel.tf"
		err := os.WriteFile(outputPath, []byte(terraformResources), 0644)
		if err != nil {
			return "", fmt.Errorf("failed to write RabbitMQ model to file: %w", err)
		}
	}

	return terraformResources, nil
}

func createRabbitMqQueueResource(pipe models.Queue, host *models.Host) string {
	resourceName := strings.ReplaceAll(pipe.Name, "-", "_")
	resource := fmt.Sprintf(`
resource "rabbitmq_queue" "%s" {
  name      = "%s"
  vhost     = "/"
  settings {
    durable    = true
    auto_delete = false
  }
}
`, resourceName, pipe.Name)

	return resource
}

func createRabbitMqTopicResource(topic models.Topic, host *models.Host) string {
	resourceName := strings.ReplaceAll(topic.Name, "-", "_")
	resource := fmt.Sprintf(`
resource "rabbitmq_exchange" "%s" {
  name  = "%s"
  vhost = "/"
  settings {
    type        = "topic"
    durable     = true
    auto_delete = false
  }
}
`, resourceName, topic.Name)

	return resource
}
