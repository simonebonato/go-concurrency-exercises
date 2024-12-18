//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

// STEPS (I THINK) I HAVE TO IMPLEMENT
// 1. add a variable called "last_updated" to the sessions	X
// 2. add a mutex to the sessions	X
// 2a. lock the session when: updating with new data	X
// 2ax. add the code to update "last_updated" when the data is updated	X
// 3. add a constant "MaxSecondsWithNoUpdate"	X
// 4. create a method for the session manager that will run in the background,
// to check the "last_updated" value and stop the session if for too long	X
// 5. add a ticker, to check only once every second (for example)	X

// What I missed: 
// In the end I had to add a mutex also for the SessionManager, since accessing the map can
// also cause a race condition
// Then another mistake was to init the mutex as *sync.Mutex, so with the memory reference,
// should have done it directly with sync.Mutex

package main

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/loong/go-concurrency-exercises/5-session-cleaner/helper"
)

const MaxSecondsWithNoUpdate int64 = 5

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sessions map[string]Session
	mu       sync.Mutex
}

// Session stores the session's data
type Session struct {
	Data         map[string]interface{}
	last_updated time.Time
	mu           sync.Mutex
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]Session),
	}

	go m.SessionCleanerRoutine()

	return m
}

func (m *SessionManager) SessionCleanerRoutine() {

	sessionTicker := time.NewTicker(1 * time.Second)
	defer sessionTicker.Stop()
	for {
		select {
		// every second the code of the ticker is run
		case <-sessionTicker.C:

			// lock the session manager
			m.mu.Lock()

			// loop through the sessions and check the elapsed time
			for sessionID, session := range m.sessions {
				session.mu.Lock()

				// see for how long they have not been updated
				not_updated_for := int64(time.Since(session.last_updated) / time.Second)

				// either delete them or unlock them if they still have time to live
				if not_updated_for >= MaxSecondsWithNoUpdate {
					fmt.Printf("\nSession with ID %s seleted after %vs with limit %v\n", sessionID, not_updated_for, MaxSecondsWithNoUpdate)
					delete(m.sessions, sessionID)
				} else {
					session.mu.Unlock()
				}
			}
			m.mu.Unlock()
		}
	}

}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := helper.MakeSessionID()
	if err != nil {
		return "", err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.sessions[sessionID] = Session{
		Data:         make(map[string]interface{}),
		last_updated: time.Now(),
	}

	return sessionID, nil
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.mu.Lock()
	session, ok := m.sessions[sessionID]
	m.mu.Unlock()

	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	_, ok := m.sessions[sessionID]

	if !ok {
		return ErrSessionNotFound
	}

	// Hint: you should renew expiry of the session here
	m.sessions[sessionID] = Session{
		Data:         data,
		last_updated: time.Now(),  // this renews the time for the session 
	}

	return nil
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}
