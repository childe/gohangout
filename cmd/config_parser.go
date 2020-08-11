package main

import (
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

func parseConfig(filename string) (map[string]interface{}, error) {
	vp := viper.New()
	vp.SetConfigFile(filename)
	err := vp.ReadInConfig()
	if err != nil {
		return nil, err
	}
	return vp.AllSettings(), nil
}

// remove sensitive info before output
func removeSensitiveInfo(config map[string]interface{}) string {
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
