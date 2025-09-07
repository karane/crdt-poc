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
		// keep element if at least one tag not removed
		for tag := range tags {
			if _, removed := s.removed[element][tag]; !removed {
				values = append(values, element)
				break
			}
		}
	}
	return values
}

// ORSetState represents the transferable state of an ORSet
type ORSetState struct {
	Added   map[string]map[string]struct{} `json:"added"`
	Removed map[string]map[string]struct{} `json:"removed"`
}

// Merge merges another ORSet state into this one
func (s *ORSet) Merge(peerState ORSetState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Merge added tags
	for element, tags := range peerState.Added {
		if s.added[element] == nil {
			s.added[element] = make(map[string]struct{})
		}
		for tag := range tags {
			s.added[element][tag] = struct{}{}
		}
	}

	// Merge removed tags
	for element, tags := range peerState.Removed {
		if s.removed[element] == nil {
			s.removed[element] = make(map[string]struct{})
		}
		for tag := range tags {
			s.removed[element][tag] = struct{}{}
		}
	}
}

func main() {
	peerURL := os.Getenv("PEER_URL")
	set := NewORSet()

	http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("element")
		if val == "" {
			http.Error(w, "Missing 'element' parameter", http.StatusBadRequest)
			return
		}
		set.Add(val)
		fmt.Fprintf(w, "Added: %s\n", val)
	})

	http.HandleFunc("/remove", func(w http.ResponseWriter, r *http.Request) {
		val := r.URL.Query().Get("element")
		if val == "" {
			http.Error(w, "Missing 'element' parameter", http.StatusBadRequest)
			return
		}
		set.Remove(val)
		fmt.Fprintf(w, "Removed: %s\n", val)
	})

	http.HandleFunc("/value", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(set.Values())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		set.mu.Lock()
		defer set.mu.Unlock()
		state := ORSetState{
			Added:   set.added,
			Removed: set.removed,
		}
		json.NewEncoder(w).Encode(state)
	})

	// New: Merge endpoint (peers can push state here)
	http.HandleFunc("/merge", func(w http.ResponseWriter, r *http.Request) {
		var peerState ORSetState
		if err := json.NewDecoder(r.Body).Decode(&peerState); err != nil {
			http.Error(w, "Invalid state payload", http.StatusBadRequest)
			return
		}
		set.Merge(peerState)
		fmt.Fprintln(w, "State merged")
	})

	// Periodically pull state from peer
	go func() {
		for {
			if peerURL == "" {
				time.Sleep(3 * time.Second)
				continue
			}
			resp, err := http.Get(peerURL + "/state")
			if err == nil {
				body, _ := ioutil.ReadAll(resp.Body)
				var peerState ORSetState
				if err := json.Unmarshal(body, &peerState); err == nil {
					set.Merge(peerState)
				}
				resp.Body.Close()
			}
			time.Sleep(4 * time.Second)
		}
	}()

	addr := ":8080"
	log.Printf("ORSet service listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
