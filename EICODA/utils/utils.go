package utils

import (
	"regexp"
	"strconv"
	"strings"
	"eicoda/models"
)

//finds a host by its name
func FindHostByName(hosts []models.Host, name string) *models.Host {
	for _, host := range hosts {
		if host.Name == name {
			return &host
		}
	}
	return nil
}

//sanitizes the name to comply with naming conventions mostly because of kubernetes
func SanitizeName(name string) string {
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	name = re.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name
}

//finds the image of a deployment artifact by name
func FindArtifactImage(artifacts []models.DeploymentArtifact, artifactName string) string {
	for _, artifact := range artifacts {
		if artifact.Name == artifactName {
			return artifact.Image
		}
	}
	return ""
}

//finds a queue by its name
func FindQueueByName(queues []models.Queue, name string) *models.Queue {
	for _, queue := range queues {
		if queue.Name == name {
			return &queue
		}
	}
	return nil
}

// finds a topic by its name
func FindTopicByName(topics []models.Topic, name string) *models.Topic {
	for _, topic := range topics {
		if topic.Name == name {
			return &topic
		}
	}
	return nil
}

//finds a filter type by its name
func FindFilterTypeByName(filterTypes []models.FilterType, name string) *models.FilterType {
	for _, filterType := range filterTypes {
		if filterType.Name == name {
			return &filterType
		}
	}
	return nil
}

//converts string values to their appropriate type
func ConvertToProperType(value string) string {
	if intValue, err := strconv.Atoi(value); err == nil {
		return strconv.Itoa(intValue)
	}
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return strconv.FormatFloat(floatValue, 'f', -1, 64)
	}
	return value
}
