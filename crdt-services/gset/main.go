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

// GSet is a grow-only set CRDT
type GSet struct {
	mu       sync.Mutex
	elements map[string]struct{}
}

func NewGSet() *GSet {
	return &GSet{
		elements: make(map[string]struct{}),
	}
}

// Add adds a new element to the set
func (s *GSet) Add(value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.elements[value] = struct{}{}
}

// Values returns all elements in the set
func (s *GSet) Values() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]string, 0, len(s.elements))
	for k := range s.elements {
		keys = append(keys, k)
	}
	return keys
}

// Merge merges another GSet into this one
func (s *GSet) Merge(other []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range other {
		s.elements[v] = struct{}{}
	}
}

func main() {
	peerURL := os.Getenv("PEER_URL")
	set := NewGSet()

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("value")
		if val == "" {
			http.Error(w, "Missing 'value' parameter", http.StatusBadRequest)
			return
		}
		set.Add(val)
		fmt.Fprintf(w, "Added: %s\n", val)
	})

	http.HandleFunc("/values", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(set.Values())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(set.Values())
	})

	// Periodically pull state from peer
	go func() {
		for {
			resp, err := http.Get(peerURL + "/state")
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				var peerState []string
				if err := json.Unmarshal(body, &peerState); err == nil {
					set.Merge(peerState)
				}
				resp.Body.Close()
			}
			time.Sleep(3 * time.Second)
		}
	}()

	addr := ":8080"
	log.Printf("GSet service listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
