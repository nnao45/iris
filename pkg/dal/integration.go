package dal

import (
	"encoding/json"

	"github.com/olegsu/iris/pkg/logger"

	"github.com/olegsu/iris/pkg/destination"
	"github.com/olegsu/iris/pkg/filter"

	"github.com/olegsu/iris/pkg/util"
	"k8s.io/api/core/v1"
)

type Integration struct {
	Name         string   `yaml:"name"`
	Filters      []string `yaml:"filters"`
	Destinations []string `yaml:"destinations"`
}

func (i *Integration) Exec(obj interface{}) (bool, error) {
	ev := obj.(*v1.Event)
	var j interface{}
	bytes, err := json.Marshal(&ev)
	if err != nil {
		return false, nil
	}
	json.Unmarshal(bytes, &j)
	result := true
	result, err = filter.IsFiltersMatched(GetDal().FilterService, i.Filters, j)
	if err != nil {
		util.EchoError(err)
		return false, err
	}
	if result == true {
		logger.Get().Info("Event pass all filter, running integration", logger.JSON{
			"name": i.Name,
		})
		destination.Exec(GetDal().DestinationService, i.Destinations, obj)
	}
	return false, nil
}
