package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/couchbaselabs/clog"
	"github.com/sbinet/liner"
)

func HandleInteractiveMode(url, prompt string) {

	// Grab HOME environment variable
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		// then try USERPROFILE for Windows
		homeDir = os.Getenv("USERPROFILE")
		if homeDir == "" {
			fmt.Printf("Unable to determine home directory, history file disabled\n")
		}
	}

	var liner = liner.NewLiner()
	defer liner.Close()

	LoadHistory(liner, homeDir)

	go signalCatcher(liner)

	for {
		line, err := liner.Prompt(prompt + "> ")
		if err != nil {
			break
		}

		if line == "" {
			continue
		}

		UpdateHistory(liner, homeDir, line)
		err = execute_internal(url, line, os.Stdout)
		if err != nil {
			clog.Error(err)
		}
	}
}

func signalCatcher(liner *liner.State) {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	liner.Close()
	os.Exit(0)
}
