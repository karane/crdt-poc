package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// ----- CRDT (RGA simplified) -----

type Operation struct {
	Type   string `json:"type"` // "insert" or "delete"
	Index  int    `json:"index"`
	Char   string `json:"char"`   // used for insert
	UserID string `json:"userId"` // identify user
}

type RGA struct {
	Text []rune
	mu   sync.Mutex
}

func NewRGA() *RGA {
	return &RGA{Text: []rune{}}
}

func (r *RGA) Apply(op Operation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch op.Type {
	case "insert":
		if op.Index >= 0 && op.Index <= len(r.Text) {
			r.Text = append(r.Text[:op.Index], append([]rune(op.Char), r.Text[op.Index:]...)...)
		}
	case "delete":
		if op.Index >= 0 && op.Index < len(r.Text) {
			r.Text = append(r.Text[:op.Index], r.Text[op.Index+1:]...)
		}
	}
}

func (r *RGA) String() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return string(r.Text)
}

// ----- WebSocket -----

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Operation)
var rga = NewRGA()
var mu sync.Mutex

func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	mu.Lock()
	clients[ws] = true
	mu.Unlock()

	// Send initial CRDT state
	initMsg, _ := json.Marshal(map[string]string{
		"type": "init",
		"text": rga.String(),
	})
	ws.WriteMessage(websocket.TextMessage, initMsg)

	for {
		var op Operation
		err := ws.ReadJSON(&op)
		if err != nil {
			log.Printf("error: %v", err)

			break
		}

		// Apply operation
		rga.Apply(op)

		// Broadcast operation
		broadcast <- op
	}

	mu.Lock()
	delete(clients, ws)
	mu.Unlock()
}

func handleMessages() {
	for {
		op := <-broadcast

		mu.Lock()
		for client := range clients {
			err := client.WriteJSON(op)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

// ----- Static HTML -----

func serveIndex(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

// ----- Main -----

func main() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
