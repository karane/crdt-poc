package main

import (
	"fmt"
	"sync"
	"time"
)

// ////////////////////////////
// G-Counter
// ////////////////////////////
type GCounter struct {
	mu     sync.Mutex
	counts map[string]int
}

func NewGCounter() *GCounter {
	return &GCounter{counts: make(map[string]int)}
}

func (c *GCounter) Increment(nodeID string, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts[nodeID] += value
}

func (c *GCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	total := 0
	for _, v := range c.counts {
		total += v
	}
	return total
}

func (c *GCounter) Merge(other *GCounter) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for node, val := range other.counts {
		if val > c.counts[node] {
			c.counts[node] = val
		}
	}
}

// ////////////////////////////
// PN-Counter
// ////////////////////////////
type PNCounter struct {
	inc *GCounter
	dec *GCounter
}

func NewPNCounter() *PNCounter {
	return &PNCounter{inc: NewGCounter(), dec: NewGCounter()}
}

func (c *PNCounter) Increment(nodeID string, value int) {
	c.inc.Increment(nodeID, value)
}

func (c *PNCounter) Decrement(nodeID string, value int) {
	c.dec.Increment(nodeID, value)
}

func (c *PNCounter) Value() int {
	return c.inc.Value() - c.dec.Value()
}

func (c *PNCounter) Merge(other *PNCounter) {
	c.inc.Merge(other.inc)
	c.dec.Merge(other.dec)
}

// ////////////////////////////
// G-Set
// ////////////////////////////
type GSet struct {
	mu    sync.Mutex
	items map[string]struct{}
}

func NewGSet() *GSet {
	return &GSet{items: make(map[string]struct{})}
}

func (s *GSet) Add(elem string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[elem] = struct{}{}
}

func (s *GSet) Contains(elem string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.items[elem]
	return ok
}

func (s *GSet) Merge(other *GSet) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for elem := range other.items {
		s.items[elem] = struct{}{}
	}
}

func (s *GSet) Elements() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := []string{}
	for elem := range s.items {
		result = append(result, elem)
	}
	return result
}

// ////////////////////////////
// LWW-Register
// ////////////////////////////
type LWWRegister struct {
	mu        sync.Mutex
	value     string
	timestamp int64
}

func (r *LWWRegister) Set(value string, timestamp int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if timestamp > r.timestamp {
		r.value = value
		r.timestamp = timestamp
	}
}

func (r *LWWRegister) Get() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.value
}

func (r *LWWRegister) Merge(other *LWWRegister) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if other.timestamp > r.timestamp {
		r.value = other.value
		r.timestamp = other.timestamp
	}
}

// ////////////////////////////
// Demo
// ////////////////////////////
func main() {
	// === Node A ===
	gA := NewGCounter()
	pnA := NewPNCounter()
	setA := NewGSet()
	regA := &LWWRegister{}

	gA.Increment("A", 5)
	pnA.Increment("A", 10)
	pnA.Decrement("A", 3)
	setA.Add("apple")
	regA.Set("hello from A", time.Now().UnixNano())

	// === Node B ===
	gB := NewGCounter()
	pnB := NewPNCounter()
	setB := NewGSet()
	regB := &LWWRegister{}

	gB.Increment("B", 7)
	pnB.Increment("B", 4)
	setB.Add("banana")
	regB.Set("hello from B", time.Now().UnixNano())

	// Simulate a delay so B’s write wins
	time.Sleep(10 * time.Millisecond)
	regB.Set("newest from B", time.Now().UnixNano())

	// === Merge A and B ===
	gA.Merge(gB)
	gB.Merge(gA)

	pnA.Merge(pnB)
	pnB.Merge(pnA)

	setA.Merge(setB)
	setB.Merge(setA)

	regA.Merge(regB)
	regB.Merge(regA)

	// === Results ===
	fmt.Println("After merging A and B:")

	fmt.Printf("GCounter A=%d, B=%d\n", gA.Value(), gB.Value())
	fmt.Printf("PNCounter A=%d, B=%d\n", pnA.Value(), pnB.Value())
	fmt.Printf("GSet A=%v, B=%v\n", setA.Elements(), setB.Elements())
	fmt.Printf("LWWRegister A=%q, B=%q\n", regA.Get(), regB.Get())
}
