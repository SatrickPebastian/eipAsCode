package main

import (
	"fmt"
	"time"
	"path/filepath"

	"eicoda/models"
	"eicoda/plugins"
	"eicoda/repositoryControllers"
	"eicoda/transformators"
)

type ApplicationController struct {
	modelParser    *ModelParser
	transformators map[string]Transformator
	plugins        map[string]Plugin
	typeController *repositoryControllers.TypeController
}

func NewApplicationController() *ApplicationController {
	return &ApplicationController{
		modelParser: NewModelParser(),
		transformators: map[string]Transformator{
			"DockerCompose": &transformators.DockerComposeTransformator{},
			"Kubernetes":    &transformators.KubernetesTransformator{},
			"RabbitMQ":      &transformators.RabbitMqTransformator{},
		},
		plugins: map[string]Plugin{
			"DockerCompose": &plugins.DockerComposePlugin{},
			"Kubernetes":    &plugins.KubernetesPlugin{},
			"Terraform":     &plugins.TerraformPlugin{},
		},
		typeController: repositoryControllers.NewTypeController(),
	}
}

//handles deployment process
func (app *ApplicationController) Deploy(path string, measure bool, noTf bool) error {
	var startTime, parseTransformTime, endTime time.Time

	if measure {
		startTime = time.Now()
	}

	fmt.Println("Starting deployment process...")
	model, err := app.modelParser.Parse(path)
	if err != nil {
		return fmt.Errorf("failed to parse model: %w", err)
	}

	//gets dir of the deployment file to pass it to transformators so that they know where to look for the critera files (if there are any)
	baseDir := filepath.Dir(path)

	fmt.Println("Transforming model...")
	if err := app.transformModel(model, baseDir, noTf); err != nil {
		return err
	}

	//measures times after parsing and transformation if flag is set
	if measure {
		parseTransformTime = time.Now()
		fmt.Printf("TIME TO PARSE AND TRANSFORM: %v\n", parseTransformTime.Sub(startTime))
	}

	fmt.Println("Executing plugins...")
	if err := app.executePlugins(model, noTf); err != nil {
		return err
	}

	fmt.Println("Successfully transformed and deployed model.")

	//measures time after entire deployment is complete (if flag is set)
	if measure {
		endTime = time.Now()
		fmt.Printf("OVERALL DEPLOYMENT TIME: %v\n", endTime.Sub(startTime))
	}

	return nil
}

//handles destruction process
func (app *ApplicationController) Destroy() error {
	fmt.Println("Starting destruction process...")

	fmt.Println("Destroying Kubernetes resources...")
	if err := app.plugins["Kubernetes"].Destroy(); err != nil {
		return fmt.Errorf("failed to destroy Kubernetes resources: %w", err)
	}

	fmt.Println("Destroying DockerCompose resources...")
	if err := app.plugins["DockerCompose"].Destroy(); err != nil {
		return fmt.Errorf("failed to destroy DockerCompose resources: %w", err)
	}

	fmt.Println("Destroying Terraform resources...")
	if err := app.plugins["Terraform"].Destroy(); err != nil {
		return fmt.Errorf("failed to destroy Terraform resources: %w", err)
	}

	fmt.Println("Successfully destroyed all resources.")
	return nil
}

func (app *ApplicationController) ProcessModel(content string) ([]string, error) {
	fmt.Println("Processing model content...")
	model, err := app.modelParser.ParseFromString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model: %w", err)
	}

	var results []string

	baseDir := filepath.Dir(".")

	fmt.Println("Transforming model with appropriate transformators...")
	for name, transformator := range app.transformators {
		transformedModel, err := transformator.Transform(model, false, baseDir)
		if err != nil {
			return nil, fmt.Errorf("failed to transform model with %s: %w", name, err)
		}
		results = append(results, fmt.Sprintf("%s Transformator:\n%s", name, transformedModel))
	}

	return results, nil
}

func (app *ApplicationController) transformModel(model *models.Model, baseDir string, noTf bool) error {
	fmt.Println("Checking if transformation for DockerCompose is needed...")
	if app.shouldTransformDockerCompose(model) {
		fmt.Println("Transforming model for DockerCompose...")
		if _, err := app.transformators["DockerCompose"].Transform(model, true, baseDir); err != nil {
			return fmt.Errorf("failed to transform model with DockerCompose: %w", err)
		}
	}
	fmt.Println("Checking if transformation for Kubernetes is needed...")
	if app.shouldTransformKubernetes(model) {
		fmt.Println("Transforming model for Kubernetes...")
		if _, err := app.transformators["Kubernetes"].Transform(model, true, baseDir); err != nil {
			return fmt.Errorf("failed to transform model with Kubernetes: %w", err)
		}
	}
	if !noTf {
		fmt.Println("Checking if transformation for RabbitMQ is needed...")
		if app.shouldTransformRabbitMQ(model) {
			fmt.Println("Transforming model for RabbitMQ...")
			if _, err := app.transformators["RabbitMQ"].Transform(model, true, baseDir); err != nil {
				return fmt.Errorf("failed to transform model with RabbitMQ: %w", err)
			}
		}
	} else {
		fmt.Println("Skipping Terraform transformations as --no-tf flag is set.")
	}
	return nil
}

func (app *ApplicationController) executePlugins(model *models.Model, noTf bool) error {

	//handles errors if anything goes seriously wrong during program execution
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("A panic occurred during plugin execution, initiating cleanup...")
			app.cleanupPlugins()
			panic(r)
		}
	}()

	if !noTf {
		fmt.Println("Executing Terraform plugin if needed...")
		if app.shouldTransformRabbitMQ(model) {
			if err := app.plugins["Terraform"].Execute(); err != nil {
				fmt.Printf("Terraform plugin execution failed: %v. Initiating cleanup...\n", err)
				app.cleanupPlugins()
				return fmt.Errorf("Terraform plugin execution failed: %w", err)
			}
		}
	} else {
		fmt.Println("Skipping Terraform plugin execution as --no-tf flag is set.")
	}
	fmt.Println("Executing Kubernetes plugin if needed...")
	if app.shouldTransformKubernetes(model) {
		if err := app.plugins["Kubernetes"].Execute(); err != nil {
			fmt.Printf("Kubernetes plugin execution failed: %v. Initiating cleanup...\n", err)
			app.cleanupPlugins()
			return fmt.Errorf("Kubernetes plugin execution failed: %w", err)
		}
	}
	fmt.Println("Executing DockerCompose plugin if needed...")
	if app.shouldTransformDockerCompose(model) {
		if err := app.plugins["DockerCompose"].Execute(); err != nil {
			fmt.Printf("DockerCompose plugin execution failed: %v. Initiating cleanup...\n", err)
			app.cleanupPlugins()
			return fmt.Errorf("DockerCompose plugin execution failed: %w", err)
		}
	}
	return nil
}

func (app *ApplicationController) shouldTransformDockerCompose(model *models.Model) bool {
	for _, host := range model.Hosts.FilterHosts {
		if host.Type == "DockerEngine" {
			for _, filter := range model.Filters {
				if filter.Host == host.Name {
					return true
				}
			}
		}
	}
	return false
}

func (app *ApplicationController) shouldTransformKubernetes(model *models.Model) bool {
	for _, host := range model.Hosts.FilterHosts {
		if host.Type == "Kubernetes" {
			for _, filter := range model.Filters {
				if filter.Host == host.Name {
					return true
				}
			}
		}
	}
	return false
}

func (app *ApplicationController) shouldTransformRabbitMQ(model *models.Model) bool {
	for _, host := range model.Hosts.PipeHosts {
		if host.Type == "RabbitMQ" {
			for _, queue := range model.Pipes.Queues {
				if queue.Host == host.Name {
					return true
				}
			}
		}
	}
	return false
}

// handle clean up if any of the plugins fail to execute the deployment
func (app *ApplicationController) cleanupPlugins() {
	fmt.Println("Starting cleanup process...")

	if err := app.plugins["Kubernetes"].Destroy(); err != nil {
		fmt.Printf("Failed to destroy Kubernetes resources during cleanup: %v\n", err)
	}

	if err := app.plugins["DockerCompose"].Destroy(); err != nil {
		fmt.Printf("Failed to destroy DockerCompose resources during cleanup: %v\n", err)
	}

	if err := app.plugins["Terraform"].Destroy(); err != nil {
		fmt.Printf("Failed to destroy Terraform resources during cleanup: %v\n", err)
	}

	fmt.Println("Cleanup process completed.")
}

//defines necessary interface of plugins
type Plugin interface {
	Execute() error
	Destroy() error
}

func (app *ApplicationController) AddType(path string) error {
	return app.typeController.AddType(path)
}
