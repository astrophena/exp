// +build windows

// syncthing-tray is a little tray utility for Syncthing on Windows.
package main

import (
	_ "embed"
	"os/exec"

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

	version, err := getVersion()
	if err != nil {
		dialog.Alert("Unable to fetch Syncthing version: %v", err)
	} else {
		info.SetTitle(version)
	}

	go func() {
		for {
			select {
			case <-gui.ClickedCh:
				if err := exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:8384").Start(); err != nil {
					dialog.Alert("Unable to open web interface: %v", err)
				}
			case <-restart.ClickedCh:
				if err := restartSyncthing(); err != nil {
					dialog.Alert("Unable to restart Syncthing: %v", err)
				}
			case <-shutdown.ClickedCh:
				if err := shutdownSyncthing(); err != nil {
					dialog.Alert("Unable to shutdown Syncthing: %v", err)
				}
			case <-quit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}
