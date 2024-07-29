package main

import (
	"fmt"

	"eicoda/models"
	"eicoda/plugins"
	"eicoda/repositoryControllers"
	"eicoda/transformators"
)

// ApplicationController is the central part of the system managing all components
type ApplicationController struct {
	modelParser    *ModelParser
	transformators map[string]Transformator
	plugins        map[string]Plugin
	typeController *repositoryControllers.TypeController
}

// NewApplicationController creates a new instance of ApplicationController
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

// Deploy handles the deployment process
func (app *ApplicationController) Deploy(path string) error {
	model, err := app.modelParser.Parse(path)
	if err != nil {
		return fmt.Errorf("failed to parse model: %w", err)
	}

	// Transform the model with the appropriate transformators
	if err := app.transformModel(model); err != nil {
		return err
	}

	// Execute the plugins based on the model
	if err := app.executePlugins(model); err != nil {
		return err
	}

	fmt.Printf("Successfully transformed and deployed model: %+v\n", model)
	return nil
}

// ProcessModel handles the transformation and returns the transformed models
func (app *ApplicationController) ProcessModel(content string) ([]string, error) {
	model, err := app.modelParser.ParseFromString(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model: %w", err)
	}

	var results []string

	// Transform the model with the appropriate transformators
	for name, transformator := range app.transformators {
		transformedModel, err := transformator.Transform(model, false)
		if err != nil {
			return nil, fmt.Errorf("failed to transform model with %s: %w", name, err)
		}
		results = append(results, fmt.Sprintf("%s Transformator:\n%s", name, transformedModel))
	}

	return results, nil
}

func (app *ApplicationController) transformModel(model *models.Model) error {
	if app.shouldTransformDockerCompose(model) {
		if _, err := app.transformators["DockerCompose"].Transform(model, true); err != nil {
			return fmt.Errorf("failed to transform model with DockerCompose: %w", err)
		}
	}
	if app.shouldTransformKubernetes(model) {
		if _, err := app.transformators["Kubernetes"].Transform(model, true); err != nil {
			return fmt.Errorf("failed to transform model with Kubernetes: %w", err)
		}
	}
	if app.shouldTransformRabbitMQ(model) {
		if _, err := app.transformators["RabbitMQ"].Transform(model, true); err != nil {
			return fmt.Errorf("failed to transform model with RabbitMQ: %w", err)
		}
	}
	return nil
}

func (app *ApplicationController) executePlugins(model *models.Model) error {
	if app.shouldTransformRabbitMQ(model) {
		if err := app.plugins["Terraform"].Execute(); err != nil {
			return fmt.Errorf("Terraform plugin execution failed: %w", err)
		}
	}
	if app.shouldTransformKubernetes(model) {
		if err := app.plugins["Kubernetes"].Execute(); err != nil {
			return fmt.Errorf("Kubernetes plugin execution failed: %w", err)
		}
	}
	if app.shouldTransformDockerCompose(model) {
		if err := app.plugins["DockerCompose"].Execute(); err != nil {
			return fmt.Errorf("DockerCompose plugin execution failed: %w", err)
		}
	}
	return nil
}

func (app *ApplicationController) shouldTransformDockerCompose(model *models.Model) bool {
	for _, host := range model.Hosts.FilterHosts {
		if host.Type == "DockerCompose" {
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

// Plugin defines the interface for all plugins
type Plugin interface {
	Execute() error
}

// AddType handles adding a new filter type
func (app *ApplicationController) AddType(path string) error {
	return app.typeController.AddType(path)
}
