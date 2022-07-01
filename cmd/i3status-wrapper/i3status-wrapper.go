/*
Command i3status-wrapper is a wrapper for the i3status command that displays
output from custom commands and shows the currently playing media title.

Usage

Simply pipe the output for i3status to i3status-wrapper and execute that
instead of i3status, for example in the i3 config file:

 bar {
   status_command i3status | i3status-wrapper
 }

i3status must be configured to output results in the i3bar JSON format:

 # ~/.config/i3status/config
 general {
   output_format = "i3bar"
 }

i3status-wrapper will run custom commands provided as arguments and add their
output before the i3status output in order:

 bar {
   status_command i3status | i3status-wrapper custom-script1.sh custom-script2.sh
 }

If your command requires arguments, then the command and arguments should be wrapped in double quotes:

 bar {
   status_command i3status | i3status-wrapper "custom-script1.sh arg1" custom-script2.sh
 }

License

Licensed under the ISC license:

  © 2022 Ilya Mateyko

  Permission to use, copy, modify, and/or distribute this software for any purpose
  with or without fee is hereby granted, provided that the above copyright notice
  and this permission notice appear in all copies.

  THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
  REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
  FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
  INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
  OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
  TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
  THIS SOFTWARE.

Forked from https://github.com/rgerardi/i3status-wrapper:

  Copyright (c) 2017 Ricardo Gerardi

  Permission is hereby granted, free of charge, to any person obtaining a copy
  of this software and associated documentation files (the "Software"), to deal
  in the Software without restriction, including without limitation the rights
  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
  copies of the Software, and to permit persons to whom the Software is
  furnished to do so, subject to the following conditions:

  The above copyright notice and this permission notice shall be included in all
  copies or substantial portions of the Software.

  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
  SOFTWARE.
*/
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.astrophena.name/exp/cmd"

	"github.com/godbus/dbus/v5"
)

// i3bar struct represents a block in the i3bar protocol
// (https://i3wm.org/docs/i3bar-protocol.html).
type i3bar struct {
	Name                string `json:"name,omitempty"`
	Instance            string `json:"instance,omitempty"`
	Markup              string `json:"markup,omitempty"`
	FullText            string `json:"full_text"` // omitting full_text produces invalid blocks
	Color               string `json:"color,omitempty"`
	ShortText           string `json:"short_text,omitempty"`
	Background          string `json:"background,omitempty"`
	Border              string `json:"border,omitempty"`
	MinWidth            int    `json:"min_width,omitempty"`
	Align               string `json:"align,omitempty"`
	Urgent              bool   `json:"urgent,omitempty"`
	Separator           bool   `json:"separator,omitempty"`
	SeparatorBlockWidth int    `json:"separator_block_width,omitempty"`
}

// i3barHeader represents the i3bar header according to the i3bar protocol.
type i3barHeader struct {
	Version     int  `json:"version"`
	StopSignal  int  `json:"stop_signal,omitempty"`
	ContSignal  int  `json:"cont_signal,omitempty"`
	ClickEvents bool `json:"click_events,omitempty"`
}

// customCommand represents a custom command to be executed.
type customCommand struct {
	command string
	args    []string
	timeout time.Duration
	result  *i3bar
	order   int // order in which the result should be displayed
}

func (c *customCommand) execute() ([]byte, error) {
	// Adding a context with timeout to handle cases of long running commands.
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	customCmd := exec.CommandContext(ctx, c.command, c.args...)
	cmdStatusOutput, err := customCmd.Output()

	// If the deadline was exceeded, just output that to the status instead of
	// failing.
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return []byte("Timed out."), nil
	}
	if err != nil {
		return nil, err
	}

	cmdStatusOutput = bytes.TrimSpace(cmdStatusOutput)
	return cmdStatusOutput, nil
}

func (c *customCommand) runJob(done chan int) {
	cmdStatusOutput, err := c.execute()
	if err != nil {
		log.Fatalf("Can't run command %q: %v", c.command, err)
	}

	// Try to parse the output as JSON with the i3bar format. If it fails
	// the output will be processed as a regular string.
	if err := json.Unmarshal(cmdStatusOutput, c.result); err != nil {
		// Not JSON, using custom fields and string output as FullText.
		c.result.Name = "customCmd"
		c.result.Instance = c.command
		c.result.FullText = string(cmdStatusOutput)
	}

	// Send status out to channel, indicates both completion and order.
	done <- c.order
}

func main() {
	cmd.SetDescription("Wrapper for the i3status command. See https://go.astrophena.name/exp/cmd/i3status-wrapper for full documentation.")
	cmd.SetArgsUsage("[commands...]")
	log.SetPrefix("i3status-wrapper: ")

	timeout := flag.Duration("timeout", 5*time.Second, "Timeout for custom command execution.")
	cmd.HandleStartup()

	bus, err := dbus.SessionBus()
	if err != nil {
		log.Fatal(err)
	}

	cmdList := make([]*customCommand, len(flag.Args()))

	for k, cmd := range flag.Args() {
		cmdSplit := strings.Split(cmd, " ")
		cmdList[k] = &customCommand{
			command: cmdSplit[0],
			args:    cmdSplit[1:],
			timeout: *timeout,
			result:  &i3bar{},
			order:   k,
		}
	}

	var (
		dec = json.NewDecoder(os.Stdin)
		enc = json.NewEncoder(os.Stdout)
	)

	// The first line is a header indicating to i3bar that JSON will be used.
	var header i3barHeader
	if err := dec.Decode(&header); err != nil {
		log.Fatalf("Can't read input: %v", err)
	}
	if err = enc.Encode(header); err != nil {
		log.Fatalf("Can't encode output JSON: %v", err)
	}
	// The second line is just the start of the endless array '['.
	t, err := dec.Token()
	if err != nil {
		log.Fatalf("Can't read input: %v", err)
	}
	fmt.Println(t)

	for dec.More() {
		// For every iteration of the loop we capture the blocks provided by i3status
		// and append custom blocks to it before sending it to i3bar.
		var blocks []*i3bar
		if err := dec.Decode(&blocks); err != nil {
			log.Fatalf("Can't decode input JSON: %v", err)
		}

		done := make(chan int)
		for _, cmd := range cmdList {
			go cmd.runJob(done)
		}

		customBlocks := make([]*i3bar, len(cmdList), len(blocks)+len(cmdList)+1)
		for i := 0; i < len(cmdList); i++ {
			d := <-done
			customBlocks[d] = cmdList[d].result
		}
		close(done)

		customBlocks = append(customBlocks, &i3bar{
			Name:     "playing",
			FullText: playing(bus),
		})
		customBlocks = append(customBlocks, blocks...)

		if err := enc.Encode(customBlocks); err != nil {
			log.Fatalf("Can't encode input JSON: %v", err)
		}

		// A comma is required to signal another entry in the array to i3bar.
		fmt.Print(",")
	}
}

// playing returns the currently playing media title.
func playing(bus *dbus.Conn) string {
	title, err := getPlayingTitle(bus)
	if err != nil {
		title = fmt.Sprintf("Error: %v", err)
	}
	if title == "" {
		return title
	}
	return "" + " " + title
}

func getPlayingTitle(bus *dbus.Conn) (string, error) {
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
