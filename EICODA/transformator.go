package main

import (
	"eicoda/models"  // Adjust the import path according to your module name
)

// Transformator defines the interface for all transformators
type Transformator interface {
	Transform(model *models.Model) error
}
