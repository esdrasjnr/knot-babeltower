package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CESARBR/knot-babeltower/pkg/logging"
	"github.com/CESARBR/knot-babeltower/pkg/user/entities"
)

// UserProxy proxy a request to the user service interface
type UserProxy interface {
	Create(user entities.User) (err error)
	CreateToken(user entities.User) (string, error)
}

// Proxy proxy a request to the user service
type Proxy struct {
	url    string
	logger logging.Logger
}

// TokenResponse represents the creating token response from the users service
type TokenResponse struct {
	Token string `json:"token"`
}

// NewUserProxy creates a proxy to the users service
func NewUserProxy(logger logging.Logger, hostname string, port uint16) *Proxy {
	url := fmt.Sprintf("http://%s:%d", hostname, port)

	logger.Debug("proxy setup to " + url)
	return &Proxy{url, logger}
}

// Create proxy the http request to user service
func (p *Proxy) Create(user entities.User) (err error) {
	p.logger.Debug("proxying request to create user")
	/**
	 * Add Timeout in http.Client to avoid blocking the request.
	 */
	client := &http.Client{Timeout: 10 * time.Second}
	jsonUser, err := json.Marshal(user)
	if err != nil {
		return err
	}

	resp, err := client.Post(p.url+"/users", "application/json", bytes.NewBuffer(jsonUser))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return p.mapErrorFromStatusCode(resp.StatusCode)
}

// CreateToken genereates a token for the specified user
func (p *Proxy) CreateToken(user entities.User) (string, error) {
	var resp *http.Response

	parsedCredentials, err := json.Marshal(user)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	resp, err = client.Post(p.url+"/tokens", "application/json", bytes.NewBuffer(parsedCredentials))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	err = p.mapErrorFromStatusCode(resp.StatusCode)
	if err != nil {
		return "", err
	}

	tr := &TokenResponse{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(tr)
	if err != nil {
		return "", nil
	}

	return tr.Token, nil
}

func (p *Proxy) mapErrorFromStatusCode(code int) error {
	var err error

	if code != http.StatusCreated {
		switch code {
		case http.StatusForbidden:
			err = entities.ErrUserForbidden
		case http.StatusConflict:
			err = entities.ErrUserExists
		}
	}

	return err
}
