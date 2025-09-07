package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

// LWWRegister is a Last-Writer-Wins CRDT
type LWWRegister struct {
	mu        sync.Mutex
	value     string
	timestamp time.Time
}

func NewLWWRegister() *LWWRegister {
	return &LWWRegister{
		timestamp: time.Time{},
	}
}

// Set updates the value with the current timestamp
func (r *LWWRegister) Set(val string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.value = val
	r.timestamp = time.Now()
}

// Value returns the current value
func (r *LWWRegister) Value() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.value
}

// Merge updates the register if the incoming timestamp is newer
func (r *LWWRegister) Merge(peerVal string, peerTs time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if peerTs.After(r.timestamp) {
		r.value = peerVal
		r.timestamp = peerTs
	}
}

func main() {
	peerURL := os.Getenv("PEER_URL")
	register := NewLWWRegister()

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("value")
		if val == "" {
			http.Error(w, "Missing 'value' parameter", http.StatusBadRequest)
			return
		}
		register.Set(val)
		fmt.Fprintf(w, "Value set to: %s\n", val)
	})

	http.HandleFunc("/value", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Value: %s\n", register.Value())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		register.mu.Lock()
		defer register.mu.Unlock()
		json.NewEncoder(w).Encode(struct {
			Value     string    `json:"value"`
			Timestamp time.Time `json:"timestamp"`
		}{
			Value:     register.value,
			Timestamp: register.timestamp,
		})
	})

	// Periodically pull state from peer
	go func() {
		for {
			resp, err := http.Get(peerURL + "/state")
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				var peerState struct {
					Value     string    `json:"value"`
					Timestamp time.Time `json:"timestamp"`
				}
				if err := json.Unmarshal(body, &peerState); err == nil {
					register.Merge(peerState.Value, peerState.Timestamp)
				}
				resp.Body.Close()
			}
			time.Sleep(4 * time.Second)
		}
	}()

	addr := ":8080"
	log.Printf("LWWRegister service listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
