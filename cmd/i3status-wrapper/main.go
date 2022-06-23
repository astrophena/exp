/*
Command i3status-wrapper is a wrapper for the i3status command. It is forked
from https://github.com/rgerardi/i3status-wrapper.

Usage

Simply pipe the output for i3status to i3status-wrapper and execute that
instead of i3status, for example in the i3 config file. i3status must be
configured to output results in the i3bar JSON format.

 i3status | i3status-wrapper

i3status-wrapper will run custom commands provided as arguments and add their
output before the i3status output in order:

 i3status | i3status-wrapper custom-script1.sh custom-script2.sh

If your command requires arguments, then the command and arguments should be wrapped in double quotes:

 i3status | i3status-wrapper "custom-script1.sh arg1" custom-script2.sh

License

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
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"go.astrophena.name/exp/cmd"
)

// i3bar struct represents a block in the i3bar protocol
// http://i3wm.org/docs/i3bar-protocol.html
type i3bar struct {
	Name                string `json:"name,omitempty"`
	Instance            string `json:"instance,omitempty"`
	Markup              string `json:"markup,omitempty"`
	FullText            string `json:"full_text,omitempty"`
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
	if ctx.Err() == context.DeadlineExceeded {
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
		fmt.Println("Cannot run command:", c.command, ":", err.Error())
		os.Exit(1)
	}

	// Here we try to parse the output as JSON with the i3bar format. If it fails
	// the output will be processed as a regular string.
	err = json.Unmarshal(cmdStatusOutput, c.result)

	if err != nil {
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

	timeout := flag.Duration("timeout", 5*time.Second, "Timeout for custom command execution.")
	cmd.HandleStartup()

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

	stdInDec := json.NewDecoder(os.Stdin)
	stdOutEnc := json.NewEncoder(os.Stdout)

	// The first line is a header indicating to i3bar that JSON will be used.
	var header i3barHeader

	err := stdInDec.Decode(&header)
	if err != nil {
		fmt.Println("Cannot read input:", err.Error())
		os.Exit(1)
	}

	err = stdOutEnc.Encode(header)
	if err != nil {
		fmt.Println("Cannot encode output json:", err.Error())
		os.Exit(1)
	}

	// The second line is just the start of the endless array '['.
	t, err := stdInDec.Token()
	if err != nil {
		fmt.Println("Cannot read input:", err.Error())
		os.Exit(1)
	}

	fmt.Println(t)

	for stdInDec.More() {
		// For every iteration of the loop we capture the blocks provided by i3status
		// and append custom blocks to it before sending it to i3bar.
		var blocks []*i3bar
		if err := stdInDec.Decode(&blocks); err != nil {
			log.Fatalf("Can't decode input JSON: %v", err.Error())
		}

		done := make(chan int)
		for _, cmd := range cmdList {
			go cmd.runJob(done)
		}

		customBlocks := make([]*i3bar, len(cmdList), len(blocks)+len(cmdList))
		for i := 0; i < len(cmdList); i++ {
			d := <-done
			customBlocks[d] = cmdList[d].result
		}
		close(done)
		customBlocks = append(customBlocks, blocks...)

		if err := stdOutEnc.Encode(customBlocks); err != nil {
			log.Fatalf("Can't encode input JSON: %v", err.Error())
		}

		// A comma is required to signal another entry in the array to i3bar.
		fmt.Print(",")
	}
}
