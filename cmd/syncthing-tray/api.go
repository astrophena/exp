// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// TODO(astrophena): Parse this from the config.
const (
	apiURL = "http://localhost:8384"
	apiKey = "Vbsh4rVTVRo3zJAiRSstYFDYXJuaWNry"
)

var client = &http.Client{}

func send(method, path string, wantStatus int) ([]byte, error) {
	req, err := http.NewRequest(method, apiURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-API-Key", apiKey)

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	slurp, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != wantStatus {
		err := fmt.Errorf("HTTP %s: %s (expected %v)", res.Status, slurp, wantStatus)
		return nil, err
	}

	return slurp, nil
}

func get200(path string) ([]byte, error)  { return send("GET", path, http.StatusOK) }
func post200(path string) ([]byte, error) { return send("POST", path, http.StatusOK) }

func getVersion() (string, error) {
	type response struct {
		Arch        string `json:"arch"`
		LongVersion string `json:"longVersion"`
		OS          string `json:"os"`
		Version     string `json:"version"`
	}

	b, err := get200("/rest/system/version")
	if err != nil {
		return "", err
	}

	res := &response{}
	if err := json.Unmarshal(b, res); err != nil {
		return "", err
	}

	return fmt.Sprintf("Syncthing %s (%s/%s)", res.Version, res.OS, res.Arch), nil
}

func restartSyncthing() error {
	_, err := post200("/rest/system/restart")
	if err != nil {
		return err
	}
	return nil
}

func shutdownSyncthing() error {
	_, err := post200("/rest/system/shutdown")
	if err != nil {
		return err
	}
	return nil
}
