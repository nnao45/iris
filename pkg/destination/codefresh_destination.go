package destination

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/olegsu/iris/pkg/logger"

	"k8s.io/api/core/v1"
)

type codefreshDestination struct {
	baseDestination `yaml:",inline"`
	Branch          string `yaml:"branch"`
	Pipeline        string `yaml:"pipeline"`
	CFToken         string `yaml:"cftoken"`
}

type codefreshPostRequestBody struct {
	Options   map[string]string `json:"options"`
	Variables map[string]string `json:"variables"`
	Contexts  []string          `json:"contexts"`
	Branch    string            `json:"branch"`
}

func (d *codefreshDestination) Exec(payload interface{}) {
	postBody := &codefreshPostRequestBody{
		Variables: make(map[string]string),
	}
	if d.Branch != "" {
		postBody.Branch = d.Branch
	}
	var ev *v1.Event
	b, _ := json.Marshal(payload)
	json.Unmarshal(b, &ev)

	postBody.Variables["IRIS_RESOURCE_NAME"] = ev.InvolvedObject.Name
	postBody.Variables["IRIS_NAMESPACE"] = ev.InvolvedObject.Namespace

	mJSON, _ := json.Marshal(postBody)
	contentReader := bytes.NewReader(mJSON)

	url := fmt.Sprintf("https://g.codefresh.io/api/pipelines/run/%s", url.QueryEscape(d.Pipeline))
	logger.Get().Info("Calling Codefresh to start pipeline", logger.JSON{
		"pipeline": d.Pipeline,
		"branch":   d.Branch,
	})
	req, _ := http.NewRequest("POST", url, contentReader)
	req.Header.Set("authorization", d.CFToken)
	req.Header.Set("User-Agent", "IRIS")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		logger.Get().Info("Codefresh success", logger.JSON{
			"BuildID": string(body),
		})
	} else {
		logger.Get().Error("Failed to execute Codefresh integration", logger.JSON{
			"message": string(body),
		})
	}
}
