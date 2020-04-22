package helpers

import (
	"encoding/json"
	"net/http"
	"bytes"
	"github.com/blindsidenetworks/mattermost-plugin-bigbluebutton/server/mattermost"
	"io/ioutil"
)

var PluginVersion string

func HttpPost(url string, requestValue interface{}, responseValue interface{}, token string) error {
	client := http.Client{}

	marshalledRequest, err := json.Marshal(requestValue)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(marshalledRequest)
	req, err := http.NewRequest(http.MethodPost, url, reader)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "bbb-mm-" + PluginVersion)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + token)

	response, err := client.Do(req)
	if err != nil {
		mattermost.API.LogError("HTTP POST ERROR: " + err.Error())
		return err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if nil != err {
		mattermost.API.LogError("HTTP POST ERROR: " + err.Error())
		return err
	}

	return json.Unmarshal(body, responseValue)
}
