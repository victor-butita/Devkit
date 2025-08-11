package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

func ConvertConfig(input, from, to string) (string, error) {
	var intermediate map[string]interface{}

	// Step 1: Unmarshal from the source format into a generic map
	switch from {
	case "json":
		if err := json.Unmarshal([]byte(input), &intermediate); err != nil {
			return "", fmt.Errorf("invalid source JSON: %w", err)
		}
	case "yaml":
		if err := yaml.Unmarshal([]byte(input), &intermediate); err != nil {
			return "", fmt.Errorf("invalid source YAML: %w", err)
		}
	case "toml":
		if err := toml.Unmarshal([]byte(input), &intermediate); err != nil {
			return "", fmt.Errorf("invalid source TOML: %w", err)
		}
	default:
		return "", fmt.Errorf("unsupported source format: %s", from)
	}

	// Step 2: Marshal from the generic map to the target format
	var output []byte
	var err error
	switch to {
	case "json":
		output, err = json.MarshalIndent(intermediate, "", "  ")
	case "yaml":
		output, err = yaml.Marshal(intermediate)
	case "toml":
		buf := new(bytes.Buffer)
		err = toml.NewEncoder(buf).Encode(intermediate)
		output = buf.Bytes()
	default:
		return "", fmt.Errorf("unsupported target format: %s", to)
	}

	if err != nil {
		return "", fmt.Errorf("error converting to target format: %w", err)
	}

	return string(output), nil
}