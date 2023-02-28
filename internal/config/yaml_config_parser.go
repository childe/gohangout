package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

type YamlParser struct{}

func (yp *YamlParser) parse(filepath string) (map[string]interface{}, error) {
	var (
		buffer []byte
		err    error
	)
	if strings.HasPrefix(filepath, "http://") || strings.HasPrefix(filepath, "https://") {
		resp, err := http.Get(filepath)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		buffer, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
	} else {
		configFile, err := os.Open(filepath)
		if err != nil {
			return nil, err
		}
		fi, _ := configFile.Stat()

		if fi.Size() == 0 {
			return nil, fmt.Errorf("config file (%s) is empty", filepath)
		}

		buffer = make([]byte, fi.Size())
		_, err = configFile.Read(buffer)
		if err != nil {
			return nil, err
		}
	}

	buffer = []byte(os.ExpandEnv(string(buffer)))

	config := make(map[string]interface{})
	err = yaml.Unmarshal(buffer, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
