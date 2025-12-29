package main

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input    string
		length   int
		expected string
	}{
		{"hello world", 5, "he..."},
		{"short", 10, "short"},
		{"exact", 5, "ex..."},
		{"", 5, ""},
		{"abc", 2, "ab"}, // Test safety fix for small widths
		{"abc", 3, "abc"},
	}

	for _, tt := range tests {
		result := truncateString(tt.input, tt.length)
		if result != tt.expected {
			t.Errorf("truncateString(%q, %d) = %q; want %q", tt.input, tt.length, result, tt.expected)
		}
	}
}

func TestAddCommas(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123", "123"},
		{"1234", "1,234"},
		{"123456", "123,456"},
		{"1234567", "1,234,567"},
		{"1234.56", "1,234.56"},
		{"-1234", "-1,234"},
		{"", ""},
	}

	for _, tt := range tests {
		result := addCommas(tt.input)
		if result != tt.expected {
			t.Errorf("addCommas(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input    float64
		decimals int
		expected string
	}{
		{1234.5678, 2, "1,234.57"},
		{1234.5, 2, "1,234.50"},
		{0, 2, "0.00"},
	}

	for _, tt := range tests {
		result := formatFloat(tt.input, tt.decimals)
		if result != tt.expected {
			t.Errorf("formatFloat(%f, %d) = %q; want %q", tt.input, tt.decimals, result, tt.expected)
		}
	}
}

func TestFormatBigFloat(t *testing.T) {
	tests := []struct {
		input    *big.Float
		decimals int
		expected string
	}{
		{big.NewFloat(1234.5678), 2, "1,234.57"},
		{nil, 2, "0"},
	}

	for _, tt := range tests {
		result := formatBigFloat(tt.input, tt.decimals)
		if result != tt.expected {
			t.Errorf("formatBigFloat(%v, %d) = %q; want %q", tt.input, tt.decimals, result, tt.expected)
		}
	}
}

func TestMasking(t *testing.T) {
	m := model{privacyMode: true}

	if got := m.maskString("100"); got != "****" {
		t.Errorf("maskString() = %q; want ****", got)
	}
	if got := m.maskAddress("0x123456"); got != "0x**...**" {
		t.Errorf("maskAddress() = %q; want 0x**...**", got)
	}

	m.privacyMode = false
	if got := m.maskString("100"); got != "100" {
		t.Errorf("maskString() = %q; want 100", got)
	}
	if got := m.maskAddress("0x123456"); got != "0x123456" {
		t.Errorf("maskAddress() = %q; want 0x123456", got)
	}
}

func TestGetFilteredTransactions(t *testing.T) {
	acc := &accountState{
		address: "0xMyAddress",
		transactions: []txInfo{
			{Hash: "0x1", From: "0xMyAddress", To: "0xOther"}, // Out
			{Hash: "0x2", From: "0xOther", To: "0xMyAddress"}, // In
			{Hash: "0x3", From: "0xOther", To: "0xOther"},     // Irrelevant
		},
	}

	m := model{}

	// Test All
	m.txFilter = "all"
	txs := m.getFilteredTransactions(acc)
	if len(txs) != 3 {
		t.Errorf("Expected 3 transactions for 'all', got %d", len(txs))
	}

	// Test In
	m.txFilter = "in"
	txs = m.getFilteredTransactions(acc)
	// Logic: !isFrom
	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions for 'in', got %d", len(txs))
	}

	// Test Out
	m.txFilter = "out"
	txs = m.getFilteredTransactions(acc)
	// Logic: isFrom
	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction for 'out', got %d", len(txs))
	}
	if len(txs) > 0 && txs[0].Hash != "0x1" {
		t.Errorf("Expected hash 0x1, got %s", txs[0].Hash)
	}
}

func TestCalculateAccountTotal(t *testing.T) {
	m := model{
		chains: []ChainConfig{
			{Name: "Ethereum", CoinGeckoID: "ethereum", Tokens: []TokenConfig{{Symbol: "USDC", CoinGeckoID: "usd-coin"}}},
		},
		prices: map[string]float64{
			"ethereum": 2000.0,
			"usd-coin": 1.0,
		},
	}

	acc := &accountState{
		balances: map[string]*big.Float{
			"Ethereum": big.NewFloat(1.5), // 1.5 * 2000 = 3000
		},
		tokenBalances: map[string]map[string]*big.Float{
			"Ethereum": {
				"USDC": big.NewFloat(100), // 100 * 1 = 100
			},
		},
	}

	total := m.calculateAccountTotal(acc)
	fTotal, _ := total.Float64()

	expected := 3100.0
	if fTotal != expected {
		t.Errorf("calculateAccountTotal = %f; want %f", fTotal, expected)
	}
}

func TestGetPrioritizedRPCs(t *testing.T) {
	m := model{
		rpcCooldowns: map[string]time.Time{
			"rpc_cooldown": time.Now().Add(time.Minute),
		},
		rpcLatencies: map[string]time.Duration{
			"rpc_fast":  10 * time.Millisecond,
			"rpc_slow":  100 * time.Millisecond,
			"rpc_error": -1,
			// rpc_unknown is missing from map
		},
	}

	input := []string{"rpc_slow", "rpc_cooldown", "rpc_error", "rpc_fast", "rpc_unknown"}

	// Expected order logic:
	// 1. Healthy (not in cooldown)
	// 2. Valid Latency (lowest first)
	// 3. Unknown Latency
	// 4. Error Latency
	// 5. Cooldown

	got := m.getPrioritizedRPCs(input)

	// Check cooldown is last
	if len(got) > 0 && got[len(got)-1] != "rpc_cooldown" {
		t.Errorf("Expected rpc_cooldown to be last, got %v", got)
	}

	// Check fast before slow
	fastIdx := -1
	slowIdx := -1
	for i, r := range got {
		if r == "rpc_fast" {
			fastIdx = i
		}
		if r == "rpc_slow" {
			slowIdx = i
		}
	}
	if fastIdx == -1 || slowIdx == -1 || fastIdx > slowIdx {
		t.Errorf("Expected rpc_fast before rpc_slow, got %v", got)
	}
}

func TestFetchChainData_Integration(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     int           `json:"id"`
			Method string        `json:"method"`
			Params []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var result interface{}
		switch req.Method {
		case "eth_getBlockByNumber":
			// Minimal header fields required by go-ethereum
			result = map[string]interface{}{
				"number":           "0x1000",
				"hash":             "0x0000000000000000000000000000000000000000000000000000000000000001",
				"parentHash":       "0x0000000000000000000000000000000000000000000000000000000000000002",
				"timestamp":        "0x5f5e1000",
				"miner":            "0x0000000000000000000000000000000000000000",
				"gasLimit":         "0x1",
				"gasUsed":          "0x0",
				"difficulty":       "0x0",
				"extraData":        "0x",
				"mixHash":          "0x0000000000000000000000000000000000000000000000000000000000000000",
				"nonce":            "0x0000000000000000",
				"stateRoot":        "0x0000000000000000000000000000000000000000000000000000000000000000",
				"receiptsRoot":     "0x0000000000000000000000000000000000000000000000000000000000000000",
				"transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
				"logsBloom":        "0x00",
			}
		case "eth_getBalance":
			// Return 2.5 ETH: 2.5 * 10^18 = 2500000000000000000
			// Hex: 0x22B1C8C1227A0000
			result = "0x22B1C8C1227A0000"
		case "eth_call":
			// Token balance. Assume 500 tokens with 6 decimals.
			// 500 * 10^6 = 500,000,000 = 0x1DCD6500
			// Padded to 32 bytes (64 hex chars)
			result = "0x000000000000000000000000000000000000000000000000000000001dcd6500"
		default:
			result = "0x0"
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Setup
	chain := ChainConfig{
		Name:    "MockChain",
		RPCURLs: []string{server.URL},
		Tokens: []TokenConfig{
			{Symbol: "TEST", Address: "0x1234567890123456789012345678901234567890", Decimals: 6},
		},
	}
	accounts := []*accountState{
		{address: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"},
	}

	// Execute
	cmd := fetchChainData(chain, accounts)
	msg := cmd()

	// Assert
	dataMsg, ok := msg.(chainDataMsg)
	if !ok {
		t.Fatalf("Expected chainDataMsg, got %T", msg)
	}

	if dataMsg.err != nil {
		t.Fatalf("fetchChainData returned error: %v", dataMsg.err)
	}

	if len(dataMsg.results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(dataMsg.results))
	}

	res := dataMsg.results[0]

	// Check Native Balance (2.5)
	expectedBal := 2.5
	gotBal, _ := res.balance.Float64()
	if gotBal != expectedBal {
		t.Errorf("Expected balance %f, got %f", expectedBal, gotBal)
	}

	// Check Token Balance (500)
	expectedToken := 500.0
	gotToken, _ := res.tokenBalances["TEST"].Float64()
	if gotToken != expectedToken {
		t.Errorf("Expected token balance %f, got %f", expectedToken, gotToken)
	}
}

func TestFetchGasPrice_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// eth_gasPrice
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x4a817c800", // 20 Gwei (20 * 10^9 = 20,000,000,000)
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	cmd := fetchGasPrice([]string{server.URL})
	msg := cmd()

	gasMsg, ok := msg.(gasPriceMsg)
	if !ok {
		t.Fatalf("Expected gasPriceMsg, got %T", msg)
	}
	if gasMsg.err != nil {
		t.Fatalf("fetchGasPrice error: %v", gasMsg.err)
	}

	// 20 Gwei = 20,000,000,000 Wei
	expected := int64(20000000000)
	if gasMsg.price.Int64() != expected {
		t.Errorf("Expected gas price %d, got %s", expected, gasMsg.price.String())
	}
}

func TestFetchEthPrice_Integration(t *testing.T) {
	// Mock Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Expect URL to contain the ID
		response := map[string]map[string]float64{
			"ethereum": {"usd": 2500.50},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Swap URL
	originalURL := coinGeckoBaseURL
	coinGeckoBaseURL = server.URL
	defer func() { coinGeckoBaseURL = originalURL }()

	// Execute
	cmd := fetchEthPrice("ethereum")
	msg := cmd()

	// Assert
	pMsg, ok := msg.(priceMsg)
	if !ok {
		t.Fatalf("Expected priceMsg, got %T", msg)
	}
	if pMsg.err != nil {
		t.Fatalf("Unexpected error: %v", pMsg.err)
	}
	if pMsg.price != 2500.50 {
		t.Errorf("Expected price 2500.50, got %f", pMsg.price)
	}
}

func TestFetchChainData_RPCError(t *testing.T) {
	// Mock Server that returns 500 Error to simulate RPC failure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	chain := ChainConfig{
		Name:    "FailChain",
		RPCURLs: []string{server.URL},
	}
	accounts := []*accountState{{address: "0x123"}}

	cmd := fetchChainData(chain, accounts)
	msg := cmd()

	dMsg, ok := msg.(chainDataMsg)
	if !ok {
		t.Fatalf("Expected chainDataMsg, got %T", msg)
	}

	if dMsg.err == nil {
		t.Error("Expected error due to RPC failure, got nil")
	}
	if len(dMsg.failedRPCs) == 0 {
		t.Error("Expected failedRPCs to be populated")
	}
}

func TestFetchChainData_RPCFailover(t *testing.T) {
	// 1. Bad Server (Simulates failure)
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer badServer.Close()

	// 2. Good Server (Simulates success)
	goodServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		var result interface{}
		switch req.Method {
		case "eth_getBlockByNumber":
			// Minimal header fields required by go-ethereum
			result = map[string]interface{}{
				"number":           "0x1000",
				"hash":             "0x0000000000000000000000000000000000000000000000000000000000000001",
				"parentHash":       "0x0000000000000000000000000000000000000000000000000000000000000002",
				"timestamp":        "0x5f5e1000",
				"miner":            "0x0000000000000000000000000000000000000000",
				"gasLimit":         "0x1",
				"gasUsed":          "0x0",
				"difficulty":       "0x0",
				"extraData":        "0x",
				"mixHash":          "0x0000000000000000000000000000000000000000000000000000000000000000",
				"nonce":            "0x0000000000000000",
				"stateRoot":        "0x0000000000000000000000000000000000000000000000000000000000000000",
				"receiptsRoot":     "0x0000000000000000000000000000000000000000000000000000000000000000",
				"transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000000",
				"logsBloom":        "0x00",
			}
		case "eth_getBalance":
			result = "0x22B1C8C1227A0000" // 2.5 ETH
		default:
			result = "0x0"
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer goodServer.Close()

	// 3. Setup Chain with Bad RPC first, then Good RPC
	chain := ChainConfig{
		Name:    "FailoverChain",
		RPCURLs: []string{badServer.URL, goodServer.URL},
	}
	accounts := []*accountState{{address: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"}}

	// 4. Execute
	cmd := fetchChainData(chain, accounts)
	msg := cmd()

	// 5. Assert
	dMsg, ok := msg.(chainDataMsg)
	if !ok {
		t.Fatalf("Expected chainDataMsg, got %T", msg)
	}

	// Should have succeeded eventually
	if len(dMsg.results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(dMsg.results))
	} else {
		val, _ := dMsg.results[0].balance.Float64()
		if val != 2.5 {
			t.Errorf("Expected balance 2.5, got %f", val)
		}
	}

	// Should have recorded the failed RPC
	if len(dMsg.failedRPCs) != 1 {
		t.Errorf("Expected 1 failed RPC, got %d", len(dMsg.failedRPCs))
	} else if dMsg.failedRPCs[0] != badServer.URL {
		t.Errorf("Expected failed RPC to be %s, got %s", badServer.URL, dMsg.failedRPCs[0])
	}
}

func TestFetchTransactions_Integration(t *testing.T) {
	targetAddress := "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"
	fromAddress := "0x1234567890123456789012345678901234567890"
	txHash := "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     int           `json:"id"`
			Method string        `json:"method"`
			Params []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		var result interface{}

		switch req.Method {
		case "eth_chainID":
			result = "0x1" // Chain ID 1
		case "eth_getBlockByNumber":
			isFull, _ := req.Params[1].(bool)
			blockHeader := map[string]interface{}{
				"number":           "0x1000",
				"hash":             "0x0000000000000000000000000000000000000000000000000000000000000001",
				"parentHash":       "0x0000000000000000000000000000000000000000000000000000000000000002",
				"timestamp":        "0x5f5e1000",
				"miner":            "0x0000000000000000000000000000000000000000",
				"gasLimit":         "0x1",
				"gasUsed":          "0x0",
				"difficulty":       "0x0",
				"extraData":        "0x",
				"mixHash":          "0x0000000000000000000000000000000000000000000000000000000000000000",
				"nonce":            "0x0000000000000000",
				"stateRoot":        "0x0000000000000000000000000000000000000000000000000000000000000000",
				"receiptsRoot":     "0x0000000000000000000000000000000000000000000000000000000000000000",
				"transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
				"logsBloom":        "0x00",
			}

			if isFull {
				// Response for client.BlockByNumber
				blockHeader["transactions"] = []map[string]interface{}{
					{
						"from":             fromAddress,
						"to":               targetAddress,
						"hash":             txHash,
						"value":            "0xde0b6b3a7640000", // 1 ETH
						"gas":              "0x5208",            // 21000
						"gasPrice":         "0x4a817c800",       // 20 Gwei
						"nonce":            "0x1",
						"blockNumber":      "0x1000",
						"blockHash":        "0x0000000000000000000000000000000000000000000000000000000000000001",
						"transactionIndex": "0x0",
						"input":            "0x",
						"v":                "0x25",
						"r":                "0x4f3a97e9a3a9f647216a200735b25695934e9b1a2063b2a63633e0de53ad8f3b",
						"s":                "0x1b959a3d30f14f8a34079b5193d96a44573139485421591d820458b38014a721",
					},
				}
				result = blockHeader
			} else {
				// Response for client.HeaderByNumber
				result = blockHeader
			}
		default:
			result = "0x0"
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Execute
	cmd := fetchTransactions(targetAddress, []string{server.URL}, 4) // 4 decimals for value
	msg := cmd()

	// Assert
	txsMsg, ok := msg.(txsMsg)
	if !ok {
		t.Fatalf("Expected txsMsg, got %T", msg)
	}

	if txsMsg.err != nil {
		t.Fatalf("fetchTransactions returned error: %v", txsMsg.err)
	}

	if len(txsMsg.txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txsMsg.txs))
	}

	tx := txsMsg.txs[0]
	if tx.Hash != txHash {
		t.Errorf("Expected hash %s, got %s", txHash, tx.Hash)
	}
	if !strings.EqualFold(tx.From, fromAddress) {
		t.Errorf("Expected from %s, got %s", fromAddress, tx.From)
	}
	if !strings.EqualFold(tx.To, targetAddress) {
		t.Errorf("Expected to %s, got %s", targetAddress, tx.To)
	}
	if tx.Value != "1.0000" {
		t.Errorf("Expected value '1.0000', got '%s'", tx.Value)
	}
	if tx.GasPrice != "20.00 Gwei" {
		t.Errorf("Expected gas price '20.00 Gwei', got '%s'", tx.GasPrice)
	}
}
