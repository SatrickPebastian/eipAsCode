package transformators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"eicoda/models"
	"eicoda/utils"
)

// RabbitMqTransformator is responsible for transforming the model to Terraform format for RabbitMQ
type RabbitMqTransformator struct{}

// Transform transforms the model to Terraform format for RabbitMQ
func (t *RabbitMqTransformator) Transform(model *models.Model) error {
	terraformResources := `
provider "rabbitmq" {
  endpoint  = "http://localhost:15672"
  username  = "guest"
  password  = "guest"
}
`

	for _, pipe := range model.Pipes.Queues {
		host := utils.FindHostByName(model.Hosts.PipeHosts, pipe.Host)
		if host != nil && host.Type == "RabbitMQ" {
			resource := createRabbitMqResource(pipe, host)
			terraformResources += resource + "\n"
		}
	}

	// Generate the file at the project root
	outputPath := filepath.Join(".", "rabbitMqModel.tf")
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create RabbitMQ model file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(terraformResources)
	if err != nil {
		return fmt.Errorf("failed to write RabbitMQ model to file: %w", err)
	}

	fmt.Printf("Successfully created RabbitMQ model file at %s\n", outputPath)
	return nil
}

func createRabbitMqResource(pipe models.Queue, host *models.Host) string {
	resourceName := strings.ReplaceAll(pipe.Name, "-", "_")
	resource := fmt.Sprintf(`
resource "rabbitmq_queue" "%s" {
  name      = "%s"
  settings {
    durable    = true
    auto_delete = false
  }
}
`, resourceName, pipe.Name)

	return resource
}
