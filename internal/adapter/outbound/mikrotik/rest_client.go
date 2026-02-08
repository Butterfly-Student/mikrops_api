package mikrotik_outbound_adapter

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"mikrops/internal/model"
)

type RestClient struct {
	baseURL  string
	username string
	password string
	useSSL   bool
	client   *http.Client
}

func NewRestClient(host string, port int, username, password string, useSSL bool) *RestClient {
	protocol := "http"
	if useSSL {
		protocol = "https"
	}

	return &RestClient{
		baseURL:  fmt.Sprintf("%s://%s:%d/rest", protocol, host, port),
		username: username,
		password: password,
		useSSL:   useSSL,
		client: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (c *RestClient) doRequest(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.username, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP error: %d - %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// PPPoE Secrets
func (c *RestClient) ListPPPoESecrets() ([]model.MikrotikPPPoESecret, error) {
	data, err := c.doRequest("GET", "/ppp/secret", nil)
	if err != nil {
		return nil, err
	}

	var secrets []model.MikrotikPPPoESecret
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, err
	}

	return secrets, nil
}

func (c *RestClient) CreatePPPoESecret(input model.MikrotikPPPoESecretInput) error {
	_, err := c.doRequest("POST", "/ppp/secret", input)
	return err
}

func (c *RestClient) UpdatePPPoESecret(id string, input model.MikrotikPPPoESecretInput) error {
	_, err := c.doRequest("PATCH", fmt.Sprintf("/ppp/secret/%s", id), input)
	return err
}

func (c *RestClient) DeletePPPoESecret(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/ppp/secret/%s", id), nil)
	return err
}

func (c *RestClient) DisablePPPoESecret(id string) error {
	_, err := c.doRequest("PATCH", fmt.Sprintf("/ppp/secret/%s", id), map[string]interface{}{"disabled": "true"})
	return err
}

func (c *RestClient) EnablePPPoESecret(id string) error {
	_, err := c.doRequest("PATCH", fmt.Sprintf("/ppp/secret/%s", id), map[string]interface{}{"disabled": "false"})
	return err
}

// PPPoE Profiles
func (c *RestClient) ListPPPoEProfiles() ([]model.MikrotikProfile, error) {
	data, err := c.doRequest("GET", "/ppp/profile", nil)
	if err != nil {
		return nil, err
	}

	var profiles []model.MikrotikProfile
	if err := json.Unmarshal(data, &profiles); err != nil {
		return nil, err
	}

	return profiles, nil
}

func (c *RestClient) CreatePPPoEProfile(input model.MikrotikProfileInput) error {
	_, err := c.doRequest("POST", "/ppp/profile", input)
	return err
}

// Hotspot Users
func (c *RestClient) ListHotspotUsers() ([]model.MikrotikHotspotUser, error) {
	data, err := c.doRequest("GET", "/ip/hotspot/user", nil)
	if err != nil {
		return nil, err
	}

	var users []model.MikrotikHotspotUser
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (c *RestClient) CreateHotspotUser(input model.MikrotikHotspotUserInput) error {
	_, err := c.doRequest("POST", "/ip/hotspot/user", input)
	return err
}

func (c *RestClient) UpdateHotspotUser(id string, input model.MikrotikHotspotUserInput) error {
	_, err := c.doRequest("PATCH", fmt.Sprintf("/ip/hotspot/user/%s", id), input)
	return err
}

func (c *RestClient) DeleteHotspotUser(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/ip/hotspot/user/%s", id), nil)
	return err
}

// Simple Queues
func (c *RestClient) ListSimpleQueues() ([]model.MikrotikSimpleQueue, error) {
	data, err := c.doRequest("GET", "/queue/simple", nil)
	if err != nil {
		return nil, err
	}

	var queues []model.MikrotikSimpleQueue
	if err := json.Unmarshal(data, &queues); err != nil {
		return nil, err
	}

	return queues, nil
}

func (c *RestClient) CreateSimpleQueue(input model.MikrotikSimpleQueueInput) error {
	_, err := c.doRequest("POST", "/queue/simple", input)
	return err
}

func (c *RestClient) UpdateSimpleQueue(id string, input model.MikrotikSimpleQueueInput) error {
	_, err := c.doRequest("PATCH", fmt.Sprintf("/queue/simple/%s", id), input)
	return err
}

func (c *RestClient) DeleteSimpleQueue(id string) error {
	_, err := c.doRequest("DELETE", fmt.Sprintf("/queue/simple/%s", id), nil)
	return err
}
