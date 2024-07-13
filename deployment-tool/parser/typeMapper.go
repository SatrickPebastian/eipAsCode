package parser

import (
    "io/ioutil"
    "gopkg.in/yaml.v2"
)

type TypeMapping struct {
    Types map[string]string `yaml:"types"`
}

// ReadTypeMapping reads typeMapping.yaml and returns the mapping
func ReadTypeMapping(filepath string) (TypeMapping, error) {
    var mapping TypeMapping
    data, err := ioutil.ReadFile(filepath)
    if err != nil {
        return mapping, err
    }
    if err := yaml.Unmarshal(data, &mapping); err != nil {
        return mapping, err
    }
    return mapping, nil
}

// WriteTypeMapping writes the updated mapping to typeMapping.yaml
func WriteTypeMapping(filepath string, mapping TypeMapping) error {
    data, err := yaml.Marshal(&mapping)
    if err != nil {
        return err
    }
    return ioutil.WriteFile(filepath, data, 0644)
}
