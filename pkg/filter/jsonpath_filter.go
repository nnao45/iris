package filter

import (
	"regexp"

	"github.com/olegsu/iris/pkg/logger"

	"github.com/yalp/jsonpath"
)

type jsonPathFilter struct {
	baseFilter `yaml:",inline"`
	Path       string `yaml:"path"`
	Value      string `yaml:"value"`
	Regexp     string `yaml:"regexp"`
	Namespace  string `yaml:"namespace"`
}

func (f *jsonPathFilter) Apply(data interface{}) (bool, error) {
	path := f.Path
	actualValue, err := jsonpath.Read(data, path)
	if err != nil {
		return false, err
	}
	if f.Value != "" {
		res := applyMatchValueFilter(f.Value, actualValue.(string))
		return res, nil
	} else if f.Regexp != "" {
		res, err := applyRegexpFilter(f.Regexp, actualValue.(string))
		if err != nil {
			return false, err
		}
		return res, nil
	} else {
		return false, nil
	}
}

func applyRegexpFilter(pattern string, value string) (bool, error) {
	match, err := regexp.MatchString(pattern, value)
	if err != nil {
		return false, err
	}
	if match == false {
		return false, nil
	}
	logger.Get().Info("JSON path match regex", logger.JSON{
		"regex":  pattern,
		"actual": value,
	})
	return true, nil
}

func applyMatchValueFilter(requiredValue string, actualValue string) bool {
	if actualValue != requiredValue {
		return false
	}
	logger.Get().Info("JSON path match", logger.JSON{
		"required": requiredValue,
		"actual":   actualValue,
	})
	return true
}
