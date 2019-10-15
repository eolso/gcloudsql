package gcloudsql

import (
	"encoding/json"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// AccessToken : Struct for storing relevant gcloud access token data
type AccessToken struct {
	lock          *sync.Mutex
	token         string
	expireTime    time.Time
	IssuedTo      string `json:"issued_to"`
	Audience      string `json:"audience"`
	UserID        string `json:"user_id"`
	Scope         string `json:"scope"`
	ExpiresIn     int    `json:"expires_in"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	AccessType    string `json:"access_type"`
}

// GenerateAccessToken : Generates an AccessToken by using the gcloud command
func GenerateAccessToken() (at AccessToken, err error) {
	var accessTokenCmd = []string{"auth", "application-default", "print-access-token"}
	output, err := exec.Command("gcloud", accessTokenCmd...).Output()

	if err != nil {
		return
	}

	at = AccessToken{
		token: strings.TrimSpace(string(output)),
	}

	requestTmpl := TemplatedHTTPRequest{
		urlText: tokenRequestURLTemplate,
		urlData: struct {
			AccessToken string
		}{
			at.token,
		},
		headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	request, err := NewHTTPRequest("GET", requestTmpl)
	if err != nil {
		return
	}

	err = ParseHTTPRequest(request, &at)

	return
}

// IsExpired : returns whether or not the AccessToken is expired
func (at AccessToken) IsExpired() bool {
	return at.expireTime.Before(time.Now())
}

func (at AccessToken) String() string {
	bytes, _ := json.MarshalIndent(at, "", "\t")

	return string(bytes)
}

func (at *AccessToken) getExpireTime() {
	at.lock.Lock()
	at.expireTime = time.Now()
	at.expireTime.Add(time.Duration(at.ExpiresIn) * time.Second)
	at.lock.Unlock()
}
