// Command playing prints a currently playing track title. Based on
// https://github.com/leberKleber/go-mpris/blob/main/player.go.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"go.astrophena.name/exp/cmd"

	"github.com/godbus/dbus/v5"
)

func main() {
	prepend := flag.String("prepend", "", "Prepend this to the track title.")
	cmd.HandleStartup()

	title, err := getTitle()
	if err != nil {
		log.Fatal(err)
	}

	var sb strings.Builder
	if *prepend != "" {
		sb.WriteString(*prepend + " ")
	}
	if title != "" {
		sb.WriteString(title)
	} else {
		sb.WriteString("Nothing is currently playing.")
	}
	fmt.Println(sb.String())
}

func getTitle() (string, error) {
	bus, err := dbus.SessionBus()
	if err != nil {
		return "", err
	}

	players, err := listPlayers(bus)
	if err != nil {
		return "", err
	}
	if len(players) == 0 {
		return "", nil
	}
	curPlayer := bus.Object(players[0], "/org/mpris/MediaPlayer2")

	metadataObj, err := curPlayer.GetProperty("org.mpris.MediaPlayer2.Player.Metadata")
	if err != nil {
		return "", err
	}
	metadata := metadataObj.Value().(map[string]dbus.Variant)

	title := metadata["xesam:title"].Value().(string)
	if title == "" || strings.Contains(title, "Yandex Music") {
		return "", nil
	}

	return title, nil
}

func listPlayers(bus *dbus.Conn) ([]string, error) {
	const prefix = "org.mpris.MediaPlayer2."

	var names, players []string

	if err := bus.BusObject().Call("org.freedesktop.DBus.ListNames", 0).Store(&names); err != nil {
		return players, err
	}

	for _, name := range names {
		if strings.HasPrefix(name, prefix) {
			players = append(players, name)
		}
	}

	return players, nil
}
