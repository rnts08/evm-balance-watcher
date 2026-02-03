package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"evmbal/pkg/watcher"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Server struct {
	watcher *watcher.Watcher
	clients map[*websocket.Conn]bool
	mu      sync.Mutex
	mux     *http.ServeMux
}

func NewServer(w *watcher.Watcher) *Server {
	s := &Server{
		watcher: w,
		clients: make(map[*websocket.Conn]bool),
		mux:     http.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("/api/status", s.handleStatus)
	s.mux.HandleFunc("/ws", s.handleWS)
}

func (s *Server) Start(port int) error {
	go s.listenToWatcher()

	fmt.Printf("API Server listening on :%d\n", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), s.mux)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"accounts": s.watcher.GetAccounts(),
		"prices":   s.watcher.GetPrices(),
	}
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
	}()

	// Send initial state
	initialData := map[string]interface{}{
		"type": "initial",
		"data": map[string]interface{}{
			"accounts": s.watcher.GetAccounts(),
			"prices":   s.watcher.GetPrices(),
		},
	}
	_ = conn.WriteJSON(initialData)

	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (s *Server) listenToWatcher() {
	sub := s.watcher.Subscribe()
	defer s.watcher.Unsubscribe(sub)

	for event := range sub {
		s.broadcast(event)
	}
}

func (s *Server) broadcast(event watcher.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		if err := client.WriteJSON(event); err != nil {
			_ = client.Close()
			delete(s.clients, client)
		}
	}
}
