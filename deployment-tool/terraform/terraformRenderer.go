package terraform

import (
    "diploy/parser"
)

type TerraformRenderer interface {
    CreateTerraformFile(config parser.Config) (string, error)
}