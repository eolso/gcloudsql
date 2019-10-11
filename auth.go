package gcloudsql

import (
	"os/exec"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type AccessToken struct {
	lock          sync.Mutex
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

func GenerateAccessToken() (AccessToken, error) {
	var accessTokenCmd = []string{"auth", "application-default", "print-access-token"}
	output, err := exec.Command("gcloud", accessTokenCmd...).Output()

	if err != nil {
		log.Fatal("Failed to get gcloud access token - ", err)
	}

	at := AccessToken{
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
		return AccessToken{}, err
	}

	err = ParseHTTPRequest(request, &at)
	if err != nil {
		return AccessToken{}, err
	}

	return at, nil
}

func (at AccessToken) IsExpired() bool {
	return at.expireTime.Before(time.Now())
}

func (at *AccessToken) getExpireTime() {
	at.lock.Lock()
	at.expireTime = time.Now()
	at.expireTime.Add(time.Duration(at.ExpiresIn) * time.Second)
	at.lock.Unlock()
}
