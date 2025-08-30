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

	"github.com/google/uuid"
)

// ORSet is an Observed-Removed Set CRDT
type ORSet struct {
	mu      sync.Mutex
	added   map[string]map[string]struct{} // element -> tag set
	removed map[string]map[string]struct{} // element -> tag set
}

func NewORSet() *ORSet {
	return &ORSet{
		added:   make(map[string]map[string]struct{}),
		removed: make(map[string]map[string]struct{}),
	}
}

// Add adds an element with a unique tag
func (s *ORSet) Add(element string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tag := uuid.New().String()
	if s.added[element] == nil {
		s.added[element] = make(map[string]struct{})
	}
	s.added[element][tag] = struct{}{}
}

// Remove removes all observed tags of an element
func (s *ORSet) Remove(element string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tags, ok := s.added[element]
	if !ok {
		return
	}
	if s.removed[element] == nil {
		s.removed[element] = make(map[string]struct{})
	}
	for tag := range tags {
		s.removed[element][tag] = struct{}{}
	}
}

// Values returns the current elements in the set
func (s *ORSet) Values() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	values := []string{}
	for element, tags := range s.added {
		for tag := range tags {
			if _, removed := s.removed[element][tag]; !removed {
				values = append(values, element)
				break
			}
		}
	}
	return values
}

// Merge merges another ORSet state into this one
func (s *ORSet) Merge(peerState map[string]map[string]struct{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for element, tags := range peerState {
		if s.added[element] == nil {
			s.added[element] = make(map[string]struct{})
		}
		for tag := range tags {
			s.added[element][tag] = struct{}{}
		}
	}
	// Note: For a full OR-Set merge, we would also need a removed map from the peer.
	// For simplicity, here we assume peerState only contains added elements.
}

func main() {
	peerURL := os.Getenv("PEER_URL")
	set := NewORSet()

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("value")
		if val == "" {
			http.Error(w, "Missing 'value' parameter", http.StatusBadRequest)
			return
		}
		set.Add(val)
		fmt.Fprintf(w, "Added: %s\n", val)
	})

	http.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("value")
		if val == "" {
			http.Error(w, "Missing 'value' parameter", http.StatusBadRequest)
			return
		}
		set.Remove(val)
		fmt.Fprintf(w, "Removed: %s\n", val)
	})

	http.HandleFunc("/values", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(set.Values())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		set.mu.Lock()
		defer set.mu.Unlock()
		json.NewEncoder(w).Encode(set.added)
	})

	// Periodically pull state from peer
	go func() {
		for {
			resp, err := http.Get(peerURL + "/state")
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				var peerState map[string]map[string]struct{}
				if err := json.Unmarshal(body, &peerState); err == nil {
					set.Merge(peerState)
				}
				resp.Body.Close()
			}
			time.Sleep(3 * time.Second)
		}
	}()

	addr := ":8080"
	log.Printf("ORSet service listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
