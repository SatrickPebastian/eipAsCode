package utils

import (
	"regexp"
	"strconv"
	"strings"
	"eicoda/models"
)

// FindHostByName finds a host by its name
func FindHostByName(hosts []models.Host, name string) *models.Host {
	for _, host := range hosts {
		if host.Name == name {
			return &host
		}
	}
	return nil
}

// SanitizeName sanitizes the name to comply with naming conventions
func SanitizeName(name string) string {
	name = strings.ToLower(name)
	re := regexp.MustCompile(`[^a-z0-9-]`)
	name = re.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	return name
}

// FindArtifactImage finds the image of a deployment artifact by name
func FindArtifactImage(artifacts []models.DeploymentArtifact, artifactName string) string {
	for _, artifact := range artifacts {
		if artifact.Name == artifactName {
			return artifact.Image
		}
	}
	return ""
}

// FindQueueByName finds a queue by its name
func FindQueueByName(queues []models.Queue, name string) *models.Queue {
	for _, queue := range queues {
		if queue.Name == name {
			return &queue
		}
	}
	return nil
}

// FindFilterTypeByName finds a filter type by its name
func FindFilterTypeByName(filterTypes []models.FilterType, name string) *models.FilterType {
	for _, filterType := range filterTypes {
		if filterType.Name == name {
			return &filterType
		}
	}
	return nil
}

// ConvertToProperType converts string values to their appropriate type
func ConvertToProperType(value string) string {
	// Try to convert to int
	if intValue, err := strconv.Atoi(value); err == nil {
		return strconv.Itoa(intValue)
	}
	// Try to convert to float
	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return strconv.FormatFloat(floatValue, 'f', -1, 64)
	}
	// Return as string if no other type matches
	return value
}
