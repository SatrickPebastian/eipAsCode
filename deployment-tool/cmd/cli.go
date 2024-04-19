package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"diploy/parser"
	"diploy/terraform"
)

var config parser.Config
var configErr error

var rootCmd = &cobra.Command{
	Use:   "diploy",
	Short: "Diploy is a CLI for deploying Enterprise Integration Patterns to different platforms",
	Long: `Diploy is a CLI tool built with Go that automates the deployment
of Enterprise Integration Patterns to different platforms using different deployment automation technologies.`,
}

var applyCmd = &cobra.Command{
    Use:   "apply",
    Short: "Apply the EIP deployment specification",
    Long:  `This command applies the Terraform configuration specified in the deployment.yaml file.`,
    RunE: func(cmd *cobra.Command, args []string) error {  // Use RunE to enable error returning
        fmt.Println("Loading YAML configuration...")
        
        if configErr != nil {
            return fmt.Errorf("Error reading YAML file: %s", configErr)
        }

        if !terraform.CheckTerraformInstalled() {
            return fmt.Errorf("Please install Terraform.")
        }

        fmt.Println("Applying Terraform configuration...")
        output, err := terraform.ApplyTerraform(config)
        if err != nil {
            fmt.Printf("Failed to apply Terraform: %s\nError: %s\n", err, output)
            return err  // Return the error to stop the process
        }

        fmt.Println("Terraform has been successfully applied:", output)
        return nil  // Return nil to indicate success
    },
}


var removeCmd = &cobra.Command{
    Use:   "remove",
    Short: "Remove the EIP deployment",
    Long:  `This command destroys the Terraform-managed infrastructure specified in the deployment.yaml file.`,
    RunE: func(cmd *cobra.Command, args []string) error {
        fmt.Println("Initializing removal process...")

        if !terraform.CheckTerraformInstalled() {
            return fmt.Errorf("Please install Terraform")
        }

        fmt.Println("Destroying Terraform-managed infrastructure...")
        output, err := terraform.DestroyTerraform(config)
        if err != nil {
            fmt.Printf("Failed to destroy Terraform-managed resources: %s\nError details: %s\n", err, output)
            return err
        }

        fmt.Println("Terraform-managed resources have been successfully destroyed.")
        fmt.Println(output)
        return nil
    },
}


func Execute() error {
	return rootCmd.Execute()
}

func init() {
	config, configErr = parser.ReadYAMLConfig("deployment.yaml")
    if configErr != nil {
        fmt.Fprintf(os.Stderr, "Error reading YAML file: %s\n", configErr)
        os.Exit(1)
    }

	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(removeCmd)
}
