// © 2021 Ilya Mateyko. All rights reserved.
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE.md file.

// +build windows

// syncthing-tray is a little tray utility for Syncthing on Windows.
package main

import (
	_ "embed"
	"encoding/xml"
	"errors"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/getlantern/systray"
	"tawesoft.co.uk/go/dialog"
)

//go:embed logo.ico
var icon []byte

func main() { systray.Run(onReady, func() {}) }

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("Syncthing")

	info := systray.AddMenuItem("Loading...", "")
	info.Disable()
	systray.AddSeparator()

	gui := systray.AddMenuItem("Open web interface", "")
	systray.AddSeparator()

	var (
		restart  = systray.AddMenuItem("Restart", "")
		shutdown = systray.AddMenuItem("Shutdown", "")
	)
	systray.AddSeparator()

	quit := systray.AddMenuItem("Quit", "")

	url, key, err := loadConfig()
	if err != nil {
		dialog.Alert("Unable to fetch URL and API key from config: %v", err)
	}
	c := newClient(url, key)

	version, err := c.version()
	if err != nil {
		dialog.Alert("Unable to fetch Syncthing version: %v", err)
	} else {
		info.SetTitle(version)
	}

	go func() {
		for {
			select {
			case <-gui.ClickedCh:
				if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", c.url).Start(); err != nil {
					dialog.Alert("Unable to open web interface: %v", err)
				}
			case <-restart.ClickedCh:
				if err := c.restart(); err != nil {
					dialog.Alert("Unable to restart Syncthing: %v", err)
				}
			case <-shutdown.ClickedCh:
				if err := c.shutdown(); err != nil {
					dialog.Alert("Unable to shutdown Syncthing: %v", err)
				}
			case <-quit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

// loadConfig fetches URL and key that is used to contact Syncthing by
// parsing it's config.
func loadConfig() (url, key string, err error) {
	// TODO(astrophena): Support non-default config paths.
	dir := os.Getenv("LocalAppData")
	if dir == "" {
		return "", "", errors.New("%LocalAppData% is not defined")
	}

	type config struct {
		XMLName xml.Name `xml:"configuration"`
		GUI     struct {
			Address string `xml:"address"`
			APIKey  string `xml:"apikey"`
		} `xml:"gui"`
	}

	b, err := os.ReadFile(filepath.Join(dir, "Syncthing", "config.xml"))
	if err != nil {
		return "", "", err
	}

	cfg := &config{}
	if err := xml.Unmarshal(b, cfg); err != nil {
		return "", "", err
	}

	return "http://" + cfg.GUI.Address, cfg.GUI.APIKey, nil
}
