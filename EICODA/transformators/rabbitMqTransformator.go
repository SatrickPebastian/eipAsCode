package transformators

import (
	"fmt"
	"os"
	"os/exec"
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
terraform {
  required_providers {
    rabbitmq = {
      source = "0UserName/rabbitmq"
      version = "1.9.1"
    }
  }
}

provider "rabbitmq" {
  endpoint  = "%s"
  username  = "%s"
  password  = "%s"
}
`

	// Assuming only one RabbitMQ host for simplicity
	var endpoint, username, password string
	for _, host := range model.Hosts.PipeHosts {
		if host.Type == "RabbitMQ" {
			endpoint = host.ConnectionString
			username = host.Username
			password = host.Password
			break
		}
	}

	terraformResources = fmt.Sprintf(terraformResources, endpoint, username, password)

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

	// Set the TF_INSECURE_SKIP_PROVIDER_VERIFY environment variable
	os.Setenv("TF_INSECURE_SKIP_PROVIDER_VERIFY", "1")

	// Initialize Terraform
	initCmd := exec.Command("terraform", "init")
	initCmd.Env = append(os.Environ(), "TF_INSECURE_SKIP_PROVIDER_VERIFY=1")
	initOutput, err := initCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to initialize Terraform: %w, output: %s", err, string(initOutput))
	}

	// Apply the Terraform configuration
	applyCmd := exec.Command("terraform", "apply", "-auto-approve")
	applyCmd.Env = append(os.Environ(), "TF_INSECURE_SKIP_PROVIDER_VERIFY=1")
	applyOutput, err := applyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply Terraform configuration: %w, output: %s", err, string(applyOutput))
	}

	fmt.Printf("Successfully applied RabbitMQ model: %s\n", string(applyOutput))
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
