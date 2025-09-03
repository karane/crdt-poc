package main

import (
	"sort"
	"sync"
)

// CharID uniquely identifies a character
type CharID struct {
	Site    string `json:"site"`
	Counter int    `json:"counter"`
}

// Character in the CRDT
type Character struct {
	ID    CharID `json:"id"`
	Value string `json:"value"`
	// Deleted flag for tombstones
	Deleted bool `json:"deleted"`
}

// Document using RGA
type Document struct {
	mu           sync.Mutex
	chars        []Character
	localCounter map[string]int
}

func NewDocument() *Document {
	return &Document{
		chars:        []Character{},
		localCounter: make(map[string]int),
	}
}

// generateID generates a new unique CharID for a site
func (d *Document) generateID(site string) CharID {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.localCounter[site]++
	return CharID{Site: site, Counter: d.localCounter[site]}
}

// Insert a character at a given position
func (d *Document) Insert(pos int, value, site string) Character {
	c := Character{
		ID:    d.generateID(site),
		Value: value,
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if pos > len(d.chars) {
		pos = len(d.chars)
	}
	d.chars = append(d.chars[:pos], append([]Character{c}, d.chars[pos:]...)...)
	return c
}

// Delete a character by ID (tombstone)
func (d *Document) Delete(id CharID) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := range d.chars {
		if d.chars[i].ID == id {
			d.chars[i].Deleted = true
			break
		}
	}
}

// Merge integrates a remote operation into the local document
func (d *Document) Merge(remote Character) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i, c := range d.chars {
		if c.ID == remote.ID {
			d.chars[i] = remote
			return
		}
	}
	d.chars = append(d.chars, remote)
	// Ensure deterministic order
	sort.Slice(d.chars, func(i, j int) bool {
		if d.chars[i].ID.Counter == d.chars[j].ID.Counter {
			return d.chars[i].ID.Site < d.chars[j].ID.Site
		}
		return d.chars[i].ID.Counter < d.chars[j].ID.Counter
	})
}

// Get visible text
func (d *Document) Text() string {
	d.mu.Lock()
	defer d.mu.Unlock()
	var result string
	for _, c := range d.chars {
		if !c.Deleted {
			result += c.Value
		}
	}
	return result
}
