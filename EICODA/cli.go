package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var appController *ApplicationController

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "myclientname",
	Short: "My Client Name is a CLI tool for deploying configurations",
	Long:  `My Client Name is a CLI tool designed to help you deploy configurations using YAML files.`,
}

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a configuration",
	Long:  `Deploy a configuration using a specified YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		if path == "" {
			fmt.Println("Path to the deployment YAML file is required.")
			return
		}
		// Use the ApplicationController to handle the deployment
		err := appController.Deploy(path)
		if err != nil {
			fmt.Printf("Deployment failed: %v\n", err)
		}
	},
}

// addTypeCmd represents the add type command
var addTypeCmd = &cobra.Command{
	Use:   "add type",
	Short: "Add a filter type",
	Long:  `Add a filter type using a specified YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		if path == "" {
			fmt.Println("Path to the filter type YAML file is required.")
			return
		}
		// Use the ApplicationController to handle adding the filter type
		err := appController.typeController.AddType(path)
		if err != nil {
			fmt.Printf("Adding filter type failed: %v\n", err)
		}
	},
}

func init() {
	appController = NewApplicationController()
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(addTypeCmd)

	// Define flags and configuration settings
	deployCmd.Flags().StringP("path", "p", "", "Path to the deployment YAML file")
	deployCmd.MarkFlagRequired("path")

	addTypeCmd.Flags().StringP("path", "p", "", "Path to the filter type YAML file")
	addTypeCmd.MarkFlagRequired("path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
