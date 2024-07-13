package target

import (
    "fmt"
    "os"
    "strings"
    "diploy/parser"
)

// ComposeRenderer is responsible for generating the docker-compose file.
type ComposeRenderer struct{}

// CreateComposeFile generates a docker-compose.yaml based on the given configuration.
func (r ComposeRenderer) CreateComposeFile(config parser.Config, mapping parser.TypeMapping) error {
    file, err := os.Create("docker-compose.yaml")
    if err != nil {
        return err
    }
    defer file.Close()

    // Start of the docker-compose file
    fmt.Fprintln(file, "version: '3.7'")
    fmt.Fprintln(file, "services:")

    // Dapr sidecar service
    fmt.Fprintln(file, "  dapr:")
    fmt.Fprintln(file, "    image: daprio/dapr:1.13.2")
    fmt.Fprintln(file, "    volumes:")
    fmt.Fprintln(file, "      - ./components:/components")
    fmt.Fprintln(file, "    command: [\"./daprd\"]")
    
    // Loop through each filter and create services with environment variables
    for _, filter := range config.Filters {
        imageName, exists := mapping.Types[filter.Type]
        if !exists {
            fmt.Printf("No image found for type %s, skipping service creation.\n", filter.Type)
            continue
        }

        fmt.Fprintf(file, "  %s:\n", filter.Name)
        fmt.Fprintf(file, "    image: %s\n", imageName)
        fmt.Fprintf(file, "    environment:\n")
        fmt.Fprintf(file, "      INPUT_QUEUE: %s\n", filter.Properties.InputQueue)
        outputQueueStr := strings.Join(filter.Properties.OutputQueues, ";")
        fmt.Fprintf(file, "      OUTPUT_QUEUES: %s\n", outputQueueStr)
        fmt.Fprintf(file, "    depends_on:\n")
        fmt.Fprintf(file, "      - dapr\n")
    }

    // Removed the explicit networks configuration
    return nil
}
