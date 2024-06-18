package utils

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitForSignal waits for termination signals and blocks until one is received.
func WaitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
