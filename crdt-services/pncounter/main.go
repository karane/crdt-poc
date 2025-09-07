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

// GCounter is a grow-only counter
type GCounter struct {
	mu     sync.Mutex
	counts map[string]int
	node   string
}

func NewGCounter(node string) *GCounter {
	return &GCounter{
		counts: make(map[string]int),
		node:   node,
	}
}

func (c *GCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[c.node]++
}

func (c *GCounter) Merge(other map[string]int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range other {
		if v > c.counts[k] {
			c.counts[k] = v
		}
	}
}

func (c *GCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	sum := 0
	for _, v := range c.counts {
		sum += v
	}
	return sum
}

// PNCounter combines two GCounters: P for increments, N for decrements
type PNCounter struct {
	P *GCounter
	N *GCounter
}

func NewPNCounter(node string) *PNCounter {
	return &PNCounter{
		P: NewGCounter(node),
		N: NewGCounter(node),
	}
}

func (c *PNCounter) Increment() {
	c.P.Increment()
}

func (c *PNCounter) Decrement() {
	c.N.Increment()
}

func (c *PNCounter) Value() int {
	return c.P.Value() - c.N.Value()
}

func (c *PNCounter) Merge(peer map[string]map[string]int) {
	if p, ok := peer["P"]; ok {
		c.P.Merge(p)
	}
	if n, ok := peer["N"]; ok {
		c.N.Merge(n)
	}
}

func main() {
	nodeID := os.Getenv("NODE_ID")
	peerURL := os.Getenv("PEER_URL")

	counter := NewPNCounter(nodeID)

	http.HandleFunc("/inc", func(w http.ResponseWriter, r *http.Request) {
		counter.Increment()
		fmt.Fprintf(w, "Incremented! Value: %d\n", counter.Value())
	})

	http.HandleFunc("/dec", func(w http.ResponseWriter, r *http.Request) {
		counter.Decrement()
		fmt.Fprintf(w, "Decremented! Value: %d\n", counter.Value())
	})

	http.HandleFunc("/value", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Value: %d\n", counter.Value())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		state := map[string]map[string]int{
			"P": counter.P.counts,
			"N": counter.N.counts,
		}
		json.NewEncoder(w).Encode(state)
	})

	// Periodically pull state from peer
	go func() {
		for {
			resp, err := http.Get(peerURL + "/state")
			// log.Printf("Ping on %s/state\n", peerURL)
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				var peerState map[string]map[string]int
				if err := json.Unmarshal(body, &peerState); err == nil {
					counter.Merge(peerState)
					// log.Printf("PNCounter merged with %s\n", peerURL)
				}
				resp.Body.Close()
			}
			time.Sleep(4 * time.Second)
		}
	}()

	addr := ":8080"
	log.Printf("PNCounter service listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
