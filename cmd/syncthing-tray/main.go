// +build windows

// syncthing-tray is a little tray utility for Syncthing on Windows.
package main

import (
	_ "embed"
	"os/exec"

	"github.com/getlantern/systray"
)

//go:embed logo.ico
var icon []byte

func main() { systray.Run(onReady, onExit) }

func onReady() {
	systray.SetIcon(icon)
	systray.SetTitle("Syncthing")

	info := systray.AddMenuItem("Loading...", "")
	info.Disable()
	systray.AddSeparator()

	gui := systray.AddMenuItem("Open web interface", "")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "")

	go func() {
		for {
			select {
			case <-gui.ClickedCh:
				exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:8384").Start()
			case <-quit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}

func onExit() {
}
