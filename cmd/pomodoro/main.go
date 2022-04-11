// Command pomodoro implements a simple Pomodoro timer.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(0)
	var (
		pomodoro   = flag.Duration("pomodoro", 25*time.Minute, "Interval of work time.")
		shortBreak = flag.Duration("short-break", 5*time.Minute, "Interval of short break time.")
		longBreak  = flag.Duration("long-break", 30*time.Minute, "Interval of long break time.")
	)
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runTimer(ctx, *pomodoro, *shortBreak, *longBreak)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-exit
	logf("Stopping...")
}

func logf(format string, args ...any) { log.Printf("==> "+format, args...) }

func runTimer(ctx context.Context, pomodoro, shortBreak, longBreak time.Duration) {
	logf("Pomodoro is %v, short break is %v, long break is %v.", pomodoro, shortBreak, longBreak)

	var pomodoros int
	wasBreak := true
	ticker := time.NewTicker(pomodoro)
	defer ticker.Stop()

loop:
	for {
		if !wasBreak {
			breakTime := shortBreak
			// Take a long break on each fourth pomodoro.
			if pomodoros%4 == 0 {
				breakTime = longBreak
			}
			ticker.Reset(breakTime)
			notifyAndLog("Pomodoro %v elapsed, time for a break of %v.", pomodoros, breakTime)
			wasBreak = true
			continue
		}

		pomodoros++
		wasBreak = false
		notifyAndLog("Pomodoro %v started for %v.", pomodoros, pomodoro)
		ticker.Reset(pomodoro)

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			break loop
		}
	}
}

func notifyAndLog(format string, args ...any) {
	if err := exec.Command("notify-send", "-u", "normal", fmt.Sprintf(format, args...)).Run(); err != nil {
		logf("Failed to send a notification: %v", err)
	}
	logf(format, args...)
}
