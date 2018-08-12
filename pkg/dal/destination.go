package dal

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"

	"k8s.io/api/core/v1"
)

type Destination struct {
	Name         string `yaml:"name"`
	URL          string `yaml:"url"`
	Secret       string `yaml:"secret"`
	Type         string `yaml:"type"`
	Branch       string `yaml:"branch"`
	Pipeline     string `yaml:"pipeline"`
	CFToken      string `yaml:"cftoken"`
	SlackToken   string `yaml:"token"`
	SlackPayload []struct {
		Key   string `yaml:"key"`
		Value string `yaml:"value"`
	} `yaml:"payload"`
}

func getHmac(secret string, payload []byte) string {
	if secret != "" {
		fmt.Println("Singing payload with secret")
		key := []byte(secret)
		mac := hmac.New(sha256.New, key)
		mac.Write(payload)
		hmac := base64.URLEncoding.EncodeToString(mac.Sum(nil))
		return hmac
	}
	return ""
}

func (d *Destination) Exec(payload interface{}) {
	fmt.Printf("Executing destination %s\n", d.Name)
	if d.Type == "" {
		execDefault(d, payload)
	} else if d.Type == "Codefresh" {
		execCodefresh(d, payload)
	} else if d.Type == "Slack" {
		execSlack(d, payload)
	}
}

func execDefault(d *Destination, payload interface{}) {
	fmt.Printf("Executing default destination to %s\n", d.URL)
	mJSON, _ := json.Marshal(payload)
	contentReader := bytes.NewReader(mJSON)
	req, _ := http.NewRequest("POST", d.URL, contentReader)
	req.Header.Set("X-IRIS-HMAC", getHmac(d.Secret, mJSON))
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	client.Do(req)
}

func execCodefresh(d *Destination, payload interface{}) {
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
	fmt.Printf("Executing Codefresh destination\n")
	fmt.Printf(string(mJSON))
	req, _ := http.NewRequest("POST", url, contentReader)
	req.Header.Set("authorization", d.CFToken)
	req.Header.Set("User-Agent", "IRIS")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		fmt.Printf("Build ID: %s\n", string(body))
	} else {
		fmt.Printf("Error:\nStatus Code: %d\nBody: %s\n", resp.StatusCode, string(body))
	}
}

func execSlack(d *Destination, payload interface{}) {
	baseURL := "https://slack.com/api/chat.postMessage"
	fmt.Printf("Executing SlackWebHook destination to %v\n", d.SlackToken)
	var slackPayload = make(map[string]interface{})
	for _, p := range d.SlackPayload {

		var tpl bytes.Buffer
		t := template.New("")
		t, _ = t.Parse(string(p.Value))
		variableSet := Interpolate(GetDal().Variables, payload)
		if err := t.Execute(&tpl, variableSet); err != nil {
			slackPayload[p.Key] = p.Value
		} else {
			slackPayload[p.Key] = tpl.String()
		}

	}
	mJSON, err := json.Marshal(slackPayload)
	if err != nil {
		fmt.Println(err)
	}
	contentReader := bytes.NewReader(mJSON)
	req, err := http.NewRequest("POST", baseURL, contentReader)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", d.SlackToken))
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf(err.Error())
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 {
		fmt.Printf("Execute Slack POST Success.\n")
	} else {
		fmt.Printf("Error:\nStatus Code: %d\nBody: %s\n", resp.StatusCode, string(body))
	}
}

type codefreshPostRequestBody struct {
	Options   map[string]string `json:"options"`
	Variables map[string]string `json:"variables"`
	Contexts  []string          `json:"contexts"`
	Branch    string            `json:"branch"`
}
