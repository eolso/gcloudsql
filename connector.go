package gcloudsql

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/briandowns/spinner"
)

// Connection : Struct for storing relevant gcloud sql connection data
type Connection struct {
	Instance    SQLInstance
	accessToken AccessToken
	httpRequest *http.Request
	response    Response
	lock        *sync.Mutex
}

// SQLInstance : Struct for storing sql relevant sql instance data
type SQLInstance struct {
	Kind            string `json:"kind"`
	State           string `json:"state"`
	DatabaseVersion string `json:"databaseVersion"`
	IPAddresses     []struct {
		Type      string `json:"type"`
		IPAddress string `json:"ipAddress"`
	} `json:"ipAddresses"`
	Project        string `json:"project"`
	SelfLink       string `json:"selfLink"`
	ConnectionName string `json:"connectionName"`
	Name           string `json:"name"`
	Region         string `json:"region"`
	GceZone        string `json:"gceZone"`
}

// Response : Struct for storing response data from gcloud sql api
type Response struct {
	Kind          string `json:"kind"`
	TargetLink    string `json:"targetLink"`
	Status        string `json:"status"`
	User          string `json:"user"`
	InsertTime    string `json:"insertTime"`
	StartTime     string `json:"startTime"`
	EndTime       string `json:"endTime"`
	OperationType string `json:"operationType"`
	Name          string `json:"name"`
	TargetID      string `json:"targetId"`
	SelfLink      string `json:"selfLink"`
	TargetProject string `json:"targetProject"`
}

// NewConnection : Creates a new Connection from a specified projectID, instanceName
func NewConnection(projectID string, instanceName string) (c Connection, err error) {
	accessToken, err := GenerateAccessToken()

	request := TemplatedHTTPRequest{
		urlText: instanceRequestURLTemplate,
		urlData: struct {
			ProjectID    string
			InstanceName string
		}{
			projectID,
			instanceName,
		},
		headers: map[string]string{
			"Authorization": "Bearer " + accessToken.token,
			"Content-Type":  "application/json",
		},
	}

	httpRequest, err := NewHTTPRequest("GET", request)
	if err != nil {
		return
	}

	var sqlInstance SQLInstance
	err = ParseHTTPRequest(httpRequest, &sqlInstance)
	if err != nil {
		return
	}

	c.Instance = sqlInstance
	c.accessToken = accessToken
	c.lock = new(sync.Mutex)

	return
}

// GetResponse : Returns the last response held by the connection
func (c Connection) GetResponse() Response {
	return c.response
}

// EnableSSL : enables the ssl required restriction on the instance
func (c *Connection) EnableSSL() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.modifySSLPolicy(true)
}

// DisableSSL : Disables the ssl required restriction on the instance
func (c *Connection) DisableSSL() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.modifySSLPolicy(false)
}

func (c *Connection) modifySSLPolicy(status bool) (err error) {
	request := TemplatedHTTPRequest{
		urlText: sslRequestURLTemplate,
		urlData: struct {
			ProjectID    string
			InstanceName string
		}{
			c.Instance.Project,
			c.Instance.Name,
		},
		headers: map[string]string{
			"Authorization": "Bearer " + c.accessToken.token,
			"Content-Type":  "application/json",
		},
		bodyText: sslRequestBodyTemplate,
		bodyData: struct {
			Value bool
		}{
			status,
		},
	}

	c.httpRequest, err = NewHTTPRequest("PATCH", request)
	if err != nil {
		return err
	}

	err = ParseHTTPRequest(c.httpRequest, &c.response)
	if err != nil {
		return err
	}

	return c.waitUntilDone()
}

// SetUserPassword : sets a specified users password
func (c *Connection) SetUserPassword(user string, password string) (err error) {
	request := TemplatedHTTPRequest{
		urlText: pwRequestURLTemplate,
		urlData: struct {
			ProjectID    string
			InstanceName string
			User         string
		}{
			c.Instance.Project,
			c.Instance.Name,
			user,
		},
		headers: map[string]string{
			"Authorization": "Bearer " + c.accessToken.token,
			"Content-Type":  "application/json",
		},
		bodyText: pwRequestBodyTemplate,
		bodyData: struct {
			User     string
			Password string
		}{
			user,
			password,
		},
	}

	c.httpRequest, err = NewHTTPRequest("PUT", request)
	if err != nil {
		return err
	}

	err = ParseHTTPRequest(c.httpRequest, &c.response)
	if err != nil {
		return err
	}

	return c.waitUntilDone()
}

func (c *Connection) waitUntilDone() (err error) {
	if c.response == (Response{}) {
		return errors.New("Connection response is empty")
	}

	request := TemplatedHTTPRequest{
		urlText: c.response.SelfLink,
		headers: map[string]string{
			"Authorization": "Bearer " + c.accessToken.token,
			"Content-Type":  "application/json",
		},
	}

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Prefix = fmt.Sprintf("Waiting for %s operation to complete ", c.response.OperationType)
	s.FinalMSG = fmt.Sprintf("%s✓\n", s.Prefix)
	s.Start()
	defer s.Stop()
	for c.response.Status != "DONE" {
		time.Sleep(1 * time.Second)

		httpRequest, err := NewHTTPRequest("GET", request)
		if err != nil {
			return err
		}

		err = ParseHTTPRequest(httpRequest, &c.response)

		if err != nil {
			return err
		}
	}

	return nil
}

func (r Response) String() string {
	bytes, _ := json.MarshalIndent(r, "", "\t")

	return string(bytes)
}
