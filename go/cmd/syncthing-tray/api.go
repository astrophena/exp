// © 2021 Ilya Mateyko. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// client is a Syncthing REST API client.
type client struct {
	url, key string
	http     *http.Client
}

// newClient returns a new Syncthing REST API client.
func newClient(url, key string) *client {
	return &client{
		url: url, key: key,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// version returns a Syncthing version.
func (c *client) version() (string, error) {
	type response struct {
		Arch        string `json:"arch"`
		LongVersion string `json:"longVersion"`
		OS          string `json:"os"`
		Version     string `json:"version"`
	}

	b, err := c.get("/rest/system/version")
	if err != nil {
		return "", err
	}

	r := &response{}
	if err := json.Unmarshal(b, r); err != nil {
		return "", err
	}

	return fmt.Sprintf("Syncthing %s (%s/%s)", r.Version, r.OS, r.Arch), nil
}

// restart immediately restarts Syncthing.
func (c *client) restart() error {
	_, err := c.post("/rest/system/restart")
	if err != nil {
		return err
	}
	return nil
}

// shutdown causes Syncthing to exit and not restart.
func (c *client) shutdown() error {
	_, err := c.post("/rest/system/shutdown")
	if err != nil {
		return err
	}
	return nil
}

func (c *client) send(method, path string, wantStatus int) ([]byte, error) {
	r, err := http.NewRequest(method, c.url+path, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Add("X-API-Key", c.key)

	rr, err := c.http.Do(r)
	if err != nil {
		return nil, err
	}

	slurp, err := io.ReadAll(rr.Body)
	if err != nil {
		return nil, err
	}
	defer rr.Body.Close()

	if rr.StatusCode != wantStatus {
		err := fmt.Errorf("HTTP %s: %s (expected %v)", rr.Status, slurp, wantStatus)
		return nil, err
	}

	return slurp, nil
}

func (c *client) get(path string) ([]byte, error) {
	return c.send(http.MethodGet, path, http.StatusOK)
}

func (c *client) post(path string) ([]byte, error) {
	return c.send(http.MethodPost, path, http.StatusOK)
}
