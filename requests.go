package gcloudsql

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"text/template"

	log "github.com/sirupsen/logrus"
)

const pwRequestURLTemplate = `https://www.googleapis.com/sql/v1beta4/projects/{{.ProjectID}}/instances/{{.InstanceName}}/users?name={{.User}}`
const pwRequestBodyTemplate = `{
	"name": "{{.User}}",
	"password": "{{.Password}}"
}`

const sslRequestURLTemplate = `https://www.googleapis.com/sql/v1beta4/projects/{{.ProjectID}}/instances/{{.InstanceName}}`
const sslRequestBodyTemplate = `{
	"settings":{
		"ipConfiguration":{
			"requireSsl":"{{.Value}}"
		}
	}
}`

const tokenRequestURLTemplate = `https://www.googleapis.com/oauth2/v1/tokeninfo?access_token={{.AccessToken}}`

const instanceRequestURLTemplate = `https://www.googleapis.com/sql/v1beta4/projects/{{.ProjectID}}/instances/{{.InstanceName}}`
const instanceRequestBodyTemplate = `{
	"settings": {
		"ipConfiguration": {
			"authorizedNetworks": [
				{{- range $index, $element := . -}}
				{{if $index}},{{end}}
				{ "value": "{{.Value}}", "name": "{{.Name}}" }
				{{- end}}
			]
		}
	}
}`

// TemplatedHTTPRequest : Struct for creating http requests through templates
type TemplatedHTTPRequest struct {
	headers map[string]string

	urlText string
	urlData interface{}

	bodyText string
	bodyData interface{}
}

// NewHTTPRequest : Creates a new *http.Request using templates
func NewHTTPRequest(method string, request TemplatedHTTPRequest) (*http.Request, error) {
	var url string

	if request.urlText != "" {
		var urlBuffer bytes.Buffer

		writer := io.Writer(&urlBuffer)
		tmpl := template.Must(template.New("url").Parse(request.urlText))

		err := tmpl.Execute(writer, request.urlData)
		if err != nil {
			log.WithFields(log.Fields{
				"Function":     "NewHTTPRequest",
				"TemplateName": tmpl.Name,
				"TemplateData": request.urlData,
			}).Error("failed to execute template")
			return nil, err
		}

		url = urlBuffer.String()
	}

	var body io.Reader
	if request.bodyText != "" {
		var bodyBuffer bytes.Buffer
		writer := io.Writer(&bodyBuffer)

		tmpl := template.Must(template.New("body").Parse(request.bodyText))
		err := tmpl.Execute(writer, request.bodyData)
		if err != nil {
			log.WithFields(log.Fields{
				"Function":     "NewHTTPRequest",
				"TemplateName": tmpl.Name,
				"TemplateData": request.urlData,
			}).Error("failed to execute template")
			return nil, err
		}

		body = bytes.NewReader(bodyBuffer.Bytes())
	}

	httpRequest, err := http.NewRequest(method, url, body)
	if err != nil {
		log.WithFields(log.Fields{
			"Function": "NewHTTPRequest",
			"Method":   method,
			"Url":      url,
			"Body":     body,
		}).Error("failed to create http request")
		return nil, err
	}

	for key, val := range request.headers {
		httpRequest.Header.Add(key, val)
	}

	return httpRequest, nil
}

// ParseHTTPRequest : Parses the response from a http request and stores the
// output in v
func ParseHTTPRequest(request *http.Request, v interface{}) error {
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.WithFields(log.Fields{
			"Function": "ParseHTTPRequest",
			"Request":  request,
		}).Error("failed to create run request")
		return err
	}

	if response.StatusCode != http.StatusOK {
		log.WithFields(log.Fields{
			"Function":   "ParseHTTPRequest",
			"StatusCode": string(response.StatusCode),
		}).Warn("request returned a non 200 response")
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(responseBody, &v)
}
