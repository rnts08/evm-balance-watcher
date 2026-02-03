package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"evmbal/pkg/config"
	"evmbal/pkg/watcher"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

func TestHandleStatus(t *testing.T) {
	w := watcher.NewWatcher(nil, nil, config.GlobalConfig{}, "")
	s := NewServer(w)

	req, _ := http.NewRequest("GET", "/api/status", nil)
	rr := httptest.NewRecorder()

	s.mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp, "accounts")
	assert.Contains(t, resp, "prices")
}

func TestHandleWS(t *testing.T) {
	w := watcher.NewWatcher(nil, nil, config.GlobalConfig{}, "")
	s := NewServer(w)
	server := httptest.NewServer(s.mux)
	defer server.Close()

	u := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	assert.NoError(t, err)
	defer func() { _ = ws.Close() }()

	// Read initial state
	var msg map[string]interface{}
	err = ws.ReadJSON(&msg)
	assert.NoError(t, err)
	assert.Equal(t, "initial", msg["type"])
}
