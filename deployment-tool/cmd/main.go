package main

import (
	"fmt"
    "diploy/parser"
    "diploy/terraform"
)

func main() {
    fmt.Println("Loading YAML configuration...")
    config, err := parser.ReadYAMLConfig("deployment.yaml")
    if err != nil {
        fmt.Printf("Error reading YAML file: %s\n", err)
        return
    }

    if !terraform.CheckTerraformInstalled() {
        fmt.Println("Please install Terraform.")
        return
    }

    renderer := terraform.RabbitMQRenderer{}
    if err := renderer.CreateTerraformFile(config); err != nil { 
        fmt.Printf("Failed to create Terraform file: %s\n", err)
        return
    }

    fmt.Println("Applying Terraform configuration...")
    if err := terraform.ApplyTerraform(); err != nil {
        fmt.Printf("Failed to apply Terraform: %s\n", err)
        return
    }

    fmt.Println("Terraform has been successfully applied.")
}
