package main

import (
    "fmt"
    "os"
    "diploy/cmd"
    "diploy/parser"
    "diploy/target"
)

func main() {
    if len(os.Args) > 1 {
        switch os.Args[1] {
        case "apply":
            // Read the configuration from the YAML file
            config, err := parser.ReadYAMLConfig("deployment.yaml")
            if err != nil {
                fmt.Printf("Error reading YAML configuration: %s\n", err)
                os.Exit(1)
            }

            // Execute apply logic
            if err := cmd.Execute(config); err != nil {
                fmt.Printf("An error occurred: %s\n", err)
                os.Exit(1)
            }

            // Only create bindings if the 'apply' command is executed successfully
            if err := target.CreateBindings(config); err != nil {
                fmt.Printf("Failed to create bindings: %s\n", err)
                os.Exit(1)
            }

			// Read type mappings before using them
            typeMapping, err := parser.ReadTypeMapping("typeMapping.yaml")
            if err != nil {
                fmt.Printf("Failed to read type mapping: %s\n", err)
                os.Exit(1)
            }

			renderer := target.ComposeRenderer{}
			if err := renderer.CreateComposeFile(config, typeMapping); err != nil {
				fmt.Printf("Failed to create docker-compose file: %s\n", err)
				os.Exit(1)
			}
            
            fmt.Println("Application setup successfully completed with apply command.")
        
        case "type":
            // Handle type mapping addition
            if len(os.Args) != 4 {
                fmt.Println("Usage: diploy type <EIPType> <DockerImage>")
                os.Exit(1)
            }
            eipType := os.Args[2]
            dockerImage := os.Args[3]
            mapping, err := parser.ReadTypeMapping("typeMapping.yaml")
            if err != nil {
                fmt.Printf("Error reading type mapping: %s\n", err)
                os.Exit(1)
            }
            if mapping.Types == nil {
                mapping.Types = make(map[string]string)
            }
            mapping.Types[eipType] = dockerImage
            if err := parser.WriteTypeMapping("typeMapping.yaml", mapping); err != nil {
                fmt.Printf("Error writing type mapping: %s\n", err)
                os.Exit(1)
            }
            fmt.Println("Type mapping updated successfully.")
        
        default:
            fmt.Println("Unknown command")
            os.Exit(1)
        }
    } else {
        fmt.Println("No command provided")
        os.Exit(1)
    }
}
