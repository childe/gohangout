package config

import (
	"errors"
	"regexp"
	"strings"

	"github.com/golang/glog"
	yaml "gopkg.in/yaml.v2"
)

type Config map[string]interface{}

type Parser interface {
	parse(filename string) (map[string]interface{}, error)
}

func ParseConfig(filename string) (map[string]interface{}, error) {
	lowerFilename := strings.ToLower(filename)
	if strings.HasSuffix(lowerFilename, ".yaml") || strings.HasSuffix(lowerFilename, ".yml") {
		yp := &YamlParser{}
		return yp.parse(filename)
	}
	return nil, errors.New("unknown config format. config filename should ends with yaml|yml")
}

// remove sensitive info before output
func RemoveSensitiveInfo(config map[string]interface{}) string {
	re := regexp.MustCompile(`(.*password:\s+)(.*)`)
	re2 := regexp.MustCompile(`(http(s)?://\w+:)\w+`)

	b, err := yaml.Marshal(config)
	if err != nil {
		glog.Errorf("marshal config error: %s", err)
		return ""
	}

	output := make([]string, 0, 0)
	for _, l := range strings.Split(string(b), "\n") {
		if re.MatchString(l) {
			output = append(output, re.ReplaceAllString(l, "${1}xxxxxx"))
			continue
		}
		if re2.MatchString(l) {
			output = append(output, re2.ReplaceAllString(l, "${1}xxxxxx"))
			continue
		}
		output = append(output, l)
	}

	return strings.Join(output, "\n")
}
