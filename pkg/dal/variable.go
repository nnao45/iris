package dal

import (
	"encoding/json"
	"fmt"

	"github.com/yalp/jsonpath"
)

type Variable struct {
	Key      string `yaml:"key"`
	Value    string `yaml:"value"`
	JSONPath string `yaml:"jsonpath"`
}

func Interpolate(vars []Variable, event interface{}) interface{} {
	p := make(map[string]string)
	var data interface{}
	bytes, err := json.Marshal(event)
	if err != nil {
		return p
	}
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		return p
	}
	for _, v := range vars {
		if v.JSONPath != "" {
			actualValue, err := jsonpath.Read(data, v.JSONPath)
			if err != nil {
				p[v.Key] = err.Error()
			} else {
				p[v.Key] = fmt.Sprintf("%s", actualValue)
			}
		} else {
			p[v.Key] = v.Value
		}
	}
	return p
}
