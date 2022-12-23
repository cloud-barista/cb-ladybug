package spider

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/cloud-barista/cb-ladybug/src/core/app"
	"github.com/go-resty/resty/v2"

	logger "github.com/sirupsen/logrus"
)

/* execute  */
func (self *Model) execute(method string, url string, body interface{}, result interface{}) (bool, error) {
	logger.Debugf("[%s] Start to execute a HTTP (url='%s')", method, url)
	resp, err := executeHTTP(method, *app.Config.SpiderUrl+url, body, result)
	if err != nil {
		return false, err
	}

	// response check
	if resp.StatusCode() == http.StatusNotFound {
		logger.Infof("[%s] Could not be found data. (url='%s')", method, url)
		return false, nil
	} else if resp.StatusCode() > 300 {
		logger.Warnf("[%s] Received error data from the Spider (statusCode=%d, url='%s', body='%v')", method, resp.StatusCode(), resp.Request.URL, resp)
		status := app.Status{}
		json.Unmarshal(resp.Body(), &status)
		// message > message 로 리턴되는 경우가 있어서 한번더 unmarshal 작업
		if json.Valid([]byte(status.Message)) {
			json.Unmarshal([]byte(status.Message), &status)
		}
		return false, errors.New(status.Message)
	}

	return true, nil
}

/* execute HTTP */
func executeHTTP(method string, url string, body interface{}, result interface{}) (*resty.Response, error) {

	req := resty.New().SetDisableWarn(true).R().SetBasicAuth(*app.Config.Username, *app.Config.Password)

	if body != nil {
		req.SetBody(body)
	}
	if result != nil {
		req.SetResult(result)
	}

	// execute
	return req.Execute(method, url)
}
