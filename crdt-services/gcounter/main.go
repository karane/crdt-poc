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

// GCounter is a simple CRDT grow-only counter
type GCounter struct {
	mu      sync.Mutex
	counts  map[string]int
	service string
}

func NewGCounter(service string) *GCounter {
	return &GCounter{
		counts:  make(map[string]int),
		service: service,
	}
}

// Increment the counter for this service
func (c *GCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[c.service]++
}

// Value returns the total
func (c *GCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	sum := 0
	for _, v := range c.counts {
		sum += v
	}
	return sum
}

// Merge updates counts with max values from another counter
func (c *GCounter) Merge(other map[string]int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range other {
		if v > c.counts[k] {
			c.counts[k] = v
		}
	}
}

func main() {
	serviceID := os.Getenv("SERVICE_ID")
	peerURL := os.Getenv("PEER_URL")
	counter := NewGCounter(serviceID)

	http.HandleFunc("/increment", func(w http.ResponseWriter, r *http.Request) {
		counter.Increment()
		fmt.Fprintf(w, "Incremented! Value: %d\n", counter.Value())
	})

	http.HandleFunc("/value", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Value: %d\n", counter.Value())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		counter.mu.Lock()
		defer counter.mu.Unlock()
		json.NewEncoder(w).Encode(counter.counts)
	})

	// Periodically pull state from peer
	go func() {
		for {
			resp, err := http.Get(peerURL + "/state")
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				peerState := make(map[string]int)
				if err := json.Unmarshal(body, &peerState); err == nil {
					counter.Merge(peerState)
				}
				resp.Body.Close()
			}
			// sync every 3s
			<-time.After(3 * time.Second)
		}
	}()

	addr := ":8080"
	log.Printf("Service %s listening on %s\n", serviceID, addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
