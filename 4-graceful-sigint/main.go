//////////////////////////////////////////////////////////////////////
//
// Given is a mock process which runs indefinitely and blocks the
// program. Right now the only way to stop the program is to send a
// SIGINT (Ctrl-C). Killing a process like that is not graceful, so we
// want to try to gracefully stop the process first.
//
// Change the program to do the following:
//   1. On SIGINT try to gracefully stop the process using
//          `proc.Stop()`
//   2. If SIGINT is called again, just kill the program (last resort)
//

package main

import (
	"grace/mockprocess"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Create a process
	proc := mockprocess.MockProcess{}

	// channel to receive os signals 
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT)

	firstTime := true

	// Run the process (blocking)

	// we send the process in a goroutine, so we can try 
	// write the stopping logic in the main part of the code
	go proc.Run()

	for {
		select {
		
		// wait for the stop signal to arrive
		case <-stopChan:
			if firstTime {
				// here try to perform the graceful shutdown
				// spoiler: it will not succeed, hence if we try again...
				firstTime = false
				go proc.Stop()
			} else {
				// ... the code will exit here
				os.Exit(1)
			}
		}
	}

}
