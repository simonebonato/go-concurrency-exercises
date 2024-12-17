//////////////////////////////////////////////////////////////////////
//
// DO NOT EDIT THIS PART
// Your task is to edit `main.go`
//

package mockserver

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup

// RunMockServer pretends to be a video processing service. It
// simulates user interacting with the Server.
func RunMockServer() {
	u1 := User{ID: 0, IsPremium: false}
	u2 := User{ID: 1, IsPremium: true}

	wg.Add(5)

	go createMockRequest(1, shortProcess, &u1)
	time.Sleep(1 * time.Second)

	go createMockRequest(2, longProcess, &u2)
	time.Sleep(2 * time.Second)

	go createMockRequest(3, shortProcess, &u1)
	time.Sleep(1 * time.Second)

	go createMockRequest(4, longProcess, &u1)
	go createMockRequest(5, shortProcess, &u2)

	wg.Wait()
}

func createMockRequest(pid int, fn func(), u *User) {
	fmt.Println("UserID:", u.ID, "\tProcess", pid, "started.")
	res := HandleRequest(fn, u)

	if res {
		fmt.Println("UserID:", u.ID, "\tProcess", pid, "done.")
	} else {
		fmt.Println("UserID:", u.ID, "\tProcess", pid, "killed. (No quota left)")
	}

	wg.Done()
}

func shortProcess() {
	time.Sleep(6 * time.Second)
}

func longProcess() {
	time.Sleep(11 * time.Second)
}

// CODE TRANSFERRED FROM MAIN FILE HERE BECAUSE IT WAS NOW RUNNIGNG OTHERWISE

const maxFreemiumTime int64 = 10

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	
	remainingTime := maxFreemiumTime - u.TimeUsed 
	if remainingTime <= 0 && !u.IsPremium {
		return false
	}

	processOver := make(chan bool)

	start := time.Now()

	go func(processOver chan bool) {
		process()
		processOver <- true	
	}(processOver)

	// if the user is not premium, start counting the time and if the 
	// remaining time goes to zero, return false
	// if he still has time, then return true, but need to check the time 
	// while the process is running in the background
	if !u.IsPremium {
		// solution proposed by ChatGPT, cool! 
		// the ticker is super cool
		tickerTime := 1 * time.Second
		ticker := time.NewTicker(tickerTime) // Check every second
		defer ticker.Stop()

		// run the select in a for loop, so it continues indefinitely
		for {

			// in the select statement, either the function is over before it
			// runs out of time, or the user does not have any more quota
			select {
				// case when the process is over
				case <- processOver:
					elapsed := int64(time.Since(start) / time.Second)
					u.TimeUsed = elapsed + u.TimeUsed
					return true
				
				// every second this case is selected
				case <- ticker.C:
					// increment the used time by the ticker interval
					u.TimeUsed += int64(tickerTime / time.Second)
					
					if u.TimeUsed >= maxFreemiumTime {

						// Quota exceeded, kill the process
						u.TimeUsed = maxFreemiumTime
						return false
					}
				}
			}
	}

	return <- processOver
}