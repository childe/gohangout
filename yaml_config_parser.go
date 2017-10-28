package main

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"
)

type YamlParser struct{}

func (yp *YamlParser) parse(filepath string) (map[string]interface{}, error) {
	config := make(map[string]interface{})
	configFile, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("Failed to open config file (%s): %s\n", filepath, err)
	}
	fi, _ := configFile.Stat()

	if fi.Size() == 0 {
		return nil, fmt.Errorf("config file (%s) is empty", filepath)
	}

	buffer := make([]byte, fi.Size())
	_, err = configFile.Read(buffer)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file (%s): %s\n", filepath, err)
	}

	err = yaml.Unmarshal(buffer, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
