package main

import (
	"eicoda/models"
)

type Transformator interface {
	Transform(model *models.Model, writeFile bool, baseDir string) (string, error)
}

type DockerComposeTransformator struct{}

func (t *DockerComposeTransformator) Transform(model *models.Model, writeFile bool, baseDir string) (string, error) {
	return "DockerCompose transformed model", nil
}

type KubernetesTransformator struct{}

func (t *KubernetesTransformator) Transform(model *models.Model, writeFile bool, baseDir string) (string, error) {
	return "Kubernetes transformed model", nil
}

type RabbitMqTransformator struct{}

func (t *RabbitMqTransformator) Transform(model *models.Model, writeFile bool, baseDir string) (string, error) {
	return "RabbitMQ transformed model", nil
}
