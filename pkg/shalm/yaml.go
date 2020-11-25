package shalm

import (
	"fmt"
	"os"

	"github.com/k14s/starlark-go/starlark"
	"gopkg.in/yaml.v2"
)

func readYamlFile(filename string, value interface{}) error {
	reader, err := os.Open(filename) // For read access.
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("Unable to open file %s for parsing: %s", filename, err.Error())
	}
	defer reader.Close()
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(value)
	if err != nil {
		return fmt.Errorf("Error during parsing file %s: %s", filename, err.Error())
	}
	return nil
}

func writeYamlFile(filename string, value interface{}) error {
	writer, err := os.Create(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("Unable to open file %s for writing: %s", filename, err.Error())
	}
	defer writer.Close()
	encoder := yaml.NewEncoder(writer)
	err = encoder.Encode(value)
	if err != nil {
		return fmt.Errorf("Error during writing file %s: %s", filename, err.Error())
	}
	return nil
}

// ReadYamlFile -
func ReadYamlFile(filename string) (starlark.Value, error) {
	var content map[string]interface{}
	err := readYamlFile(filename, &content)
	if err != nil {
		return starlark.None, err
	}
	return WrapDict(ToStarlark(content)), nil
}

// WriteYamlFile -
func WriteYamlFile(filename string, value starlark.Value) error {
	return writeYamlFile(filename, toGo(value))
}
