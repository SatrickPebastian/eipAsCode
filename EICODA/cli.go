package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var appController *ApplicationController

//when called without any commands
var rootCmd = &cobra.Command{
	Use:   "eicoda",
	Short: "EICODA is a CLI tool for deploying pipes and filters architectures",
	Long:  `EICODA is a CLI tool designed to help you deploy pipes and filters architecture configurations using YAML files.`,
}

//start deploy process
var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy a configuration",
	Long:  `Deploy a configuration using a specified YAML file.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		measure, _ := cmd.Flags().GetBool("measure")
		noTf, _ := cmd.Flags().GetBool("no-tf")
		if path == "" {
			fmt.Println("Path to the deployment YAML file is required.")
			return
		}
		
		err := appController.Deploy(path, measure, noTf)
		if err != nil {
			fmt.Printf("Deployment failed: %v\n", err)
		}
	},
}

//add types to persist them in /repositoryControllers/types.yaml or /repositoryControllers/hostTypes.yaml files
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
		
		err := appController.typeController.AddType(path)
		if err != nil {
			fmt.Printf("Adding filter type failed: %v\n", err)
		}
	},
}

var processCmd = &cobra.Command{
	Use:   "process",
	Short: "Process a deployment model",
	Long:  `Process a deployment model and return the transformed models.`,
	Run: func(cmd *cobra.Command, args []string) {
		content, _ := cmd.Flags().GetString("content")
		if content == "" {
			fmt.Println("Content of the deployment model is required.")
			return
		}
		results, err := appController.ProcessModel(content)
		if err != nil {
			fmt.Printf("Processing failed: %v\n", err)
			return
		}
		for _, result := range results {
			fmt.Println(result)
		}
	},
}

//destroy deployed resources. Uses files that are in kubernetesModel.yaml, rabbitMqModel.yaml and docker-compose.yaml relative to the eicoda binary
var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy a deployment",
	Long:  `Destroy a deployment that was previously created.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := appController.Destroy()
		if err != nil {
			fmt.Printf("Destroy failed: %v\n", err)
		} else {
			fmt.Println("Deployment successfully destroyed.")
		}
	},
}

func init() {
	appController = NewApplicationController()
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(addTypeCmd)
	rootCmd.AddCommand(processCmd)
	rootCmd.AddCommand(destroyCmd)

	deployCmd.Flags().StringP("path", "p", "", "Path to the deployment YAML file")
	deployCmd.MarkFlagRequired("path")
	deployCmd.Flags().BoolP("measure", "m", false, "Measure the deployment performance")
	deployCmd.Flags().Bool("no-tf", false, "Skip Terraform-related actions during deployment")

	addTypeCmd.Flags().StringP("path", "p", "", "Path to the filter type YAML file")
	addTypeCmd.MarkFlagRequired("path")

	processCmd.Flags().StringP("content", "c", "", "Content of the deployment YAML file")
	processCmd.MarkFlagRequired("content")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
