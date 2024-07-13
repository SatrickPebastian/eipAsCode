package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
    "diploy/parser"
    "diploy/terraform"
	//"diploy/target"
)

// Command structures might need a way to hold configuration
type commandContext struct {
    config parser.Config
}

var rootCmd = &cobra.Command{
    Use:   "diploy",
    Short: "Diploy is a CLI for deploying Enterprise Integration Patterns to different platforms",
    Long: `Diploy is a CLI tool built with Go that automates the deployment
of Enterprise Integration Patterns to different platforms using different deployment automation technologies.`,
}

// Prepare the command context to hold configuration
var ctx commandContext

var applyCmd = &cobra.Command{
    Use:   "apply",
    Short: "Apply the EIP deployment specification",
    Long:  `This command applies the Terraform configuration specified in the deployment.yaml file.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return applyTerraform(ctx.config) // Access the configuration from the context
    },
}

var removeCmd = &cobra.Command{
    Use:   "remove",
    Short: "Remove the EIP deployment",
    Long:  `This command destroys the Terraform-managed infrastructure specified in the deployment.yaml file.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        return removeTerraform(ctx.config) // Access the configuration from the context
    },
}

// var dockerUpCmd = &cobra.Command{
//     Use:   "docker-up",
//     Short: "Start Docker Compose services",
//     Long:  `This command starts the services defined in the docker-compose.yaml file.`,
//     RunE: func(cmd *cobra.Command, args []string) error {
//         if !CheckDockerComposeInstalled() {
//             return fmt.Errorf("Please install Docker Compose.")
//         }

//         output, err := DockerComposeUp()
//         if err != nil {
//             fmt.Printf("Failed to start Docker Compose services: %s\nError: %s\n", err, output)
//             return err
//         }

//         fmt.Println("Docker Compose services started successfully.")
//         return nil
//     },
// }

// var dockerDownCmd = &cobra.Command{
//     Use:   "docker-down",
//     Short: "Stop Docker Compose services",
//     Long:  `This command stops the services defined in the docker-compose.yaml file.`,
//     RunE: func(cmd *cobra.Command, args []string) error {
//         if !CheckDockerComposeInstalled() {
//             return fmt.Errorf("Please install Docker Compose")
//         }

//         output, err := DockerComposeDown()
//         if err != nil {
//             fmt.Printf("Failed to stop Docker Compose services: %s\nError details: %s\n", err, output)
//             return err
//         }

//         fmt.Println("Docker Compose services have been successfully stopped.")
//         return nil
//     },
// }

// Execute runs the root command and hence, the entire CLI application
func Execute(config parser.Config) error {
    // Set the config in our context
    ctx.config = config

    rootCmd.AddCommand(applyCmd)
    rootCmd.AddCommand(removeCmd)
	//rootCmd.AddCommand(dockerUpCmd)
    //rootCmd.AddCommand(dockerDownCmd)

    return rootCmd.Execute()
}

// apply performs the application of resources
func applyTerraform(config parser.Config) error {
    if !terraform.CheckTerraformInstalled() {
        return fmt.Errorf("Please install Terraform.")
    }

    output, err := terraform.ApplyTerraform(config)
    if err != nil {
        fmt.Printf("Failed to apply Terraform: %s\nError: %s\n", err, output)
        return err
    }

    fmt.Println("Terraform has been successfully applied:", output)
    return nil
}

// remove performs the removal of resources
func removeTerraform(config parser.Config) error {
    if !terraform.CheckTerraformInstalled() {
        return fmt.Errorf("Please install Terraform")
    }

    output, err := terraform.DestroyTerraform(config)
    if err != nil {
        fmt.Printf("Failed to destroy Terraform-managed resources: %s\nError details: %s\n", err, output)
        return err
    }

    fmt.Println("Terraform-managed resources have been successfully destroyed.")
    fmt.Println(output)
    return nil
}

// func applyCompose() error {
//     if !CheckDockerComposeInstalled() {
//         return fmt.Errorf("Docker Compose is not installed. Please install Docker Compose.")
//     }

//     output, err := DockerComposeUp()
//     if err != nil {
//         fmt.Printf("Failed to start Docker Compose services: %s\nError: %s\n", err, output)
//         return err
//     }

//     fmt.Println("Docker Compose services started successfully:", output)
//     return nil
// }

// // removeCompose stops and removes Docker Compose services.
// func removeCompose() error {
//     if !CheckDockerComposeInstalled() {
//         return fmt.Errorf("Docker Compose is not installed. Please install Docker Compose")
//     }

//     output, err := DockerComposeDown()
//     if err != nil {
//         fmt.Printf("Failed to stop Docker Compose services: %s\nError details: %s\n", err, output)
//         return err
//     }

//     fmt.Println("Docker Compose services have been successfully stopped.")
//     fmt.Println(output)
//     return nil
// }
