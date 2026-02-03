package rpc

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"evmbal/pkg/config"
	"evmbal/pkg/models"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

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
			result = map[string]interface{}{
				"number":           "0x1000",
				"hash":             "0x0000000000000000000000000000000000000000000000000000000000000001",
				"parentHash":       "0x0000000000000000000000000000000000000000000000000000000000000002",
				"sha3Uncles":       "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
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
				"transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000001",
				"logsBloom":        "0x" + strings.Repeat("00", 256),
			}
		case "eth_getBalance":
			result = "0x22B1C8C1227A0000"
		case "eth_call":
			result = "0x000000000000000000000000000000000000000000000000000000001dcd6500"
		default:
			result = "0x0"
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	chain := config.ChainConfig{
		Name:    "MockChain",
		RPCURLs: []string{server.URL},
		Tokens: []config.TokenConfig{
			{Symbol: "TEST", Address: "0x1234567890123456789012345678901234567890", Decimals: 6},
		},
	}
	accounts := []*models.Account{
		{Address: "0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B"},
	}

	dataMsg, err := FetchChainData(chain, accounts)
	if err != nil {
		t.Fatalf("FetchChainData returned error: %v", err)
	}

	if len(dataMsg.Results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(dataMsg.Results))
	}

	res := dataMsg.Results[0]
	expectedBal := 2.5
	gotBal, _ := res.Balance.Float64()
	if gotBal != expectedBal {
		t.Errorf("Expected balance %f, got %f", expectedBal, gotBal)
	}

	expectedToken := 500.0
	gotToken, _ := res.TokenBalances["TEST"].Float64()
	if gotToken != expectedToken {
		t.Errorf("Expected token balance %f, got %f", expectedToken, gotToken)
	}
}

func TestFetchGasPrice_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "0x4a817c800", // 20 Gwei
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	gasMsg, err := FetchGasPrice([]string{server.URL})
	if err != nil {
		t.Fatalf("FetchGasPrice error: %v", err)
	}

	expected := int64(20000000000)
	if gasMsg.Price.Int64() != expected {
		t.Errorf("Expected gas price %d, got %s", expected, gasMsg.Price.String())
	}
}

func TestFetchEthPrice_Integration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]map[string]float64{
			"ethereum": {"usd": 2500.50},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	originalURL := CoinGeckoBaseURL
	CoinGeckoBaseURL = server.URL
	defer func() { CoinGeckoBaseURL = originalURL }()

	pMsg, err := FetchEthPrice("ethereum")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pMsg.Price != 2500.50 {
		t.Errorf("Expected price 2500.50, got %f", pMsg.Price)
	}
}

func TestFetchTransactions_Integration(t *testing.T) {
	key, _ := crypto.GenerateKey()
	fromAddr := crypto.PubkeyToAddress(key.PublicKey)
	fromAddress := fromAddr.Hex()

	targetAddr := common.HexToAddress("0xAb5801a7D398351b8bE11C439e05C5B3259aeC9B")
	targetAddress := targetAddr.Hex()

	txData := types.NewTransaction(
		1,
		targetAddr,
		big.NewInt(1000000000000000000),
		21000,
		big.NewInt(20000000000),
		nil,
	)

	signer := types.NewLondonSigner(big.NewInt(1))
	signedTx, err := types.SignTx(txData, signer, key)
	if err != nil {
		t.Fatal(err)
	}

	sigV, sigR, sigS := signedTx.RawSignatureValues()
	txHash := signedTx.Hash().Hex()

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
		case "eth_chainId":
			result = "0x1"
		case "eth_getBlockByNumber":
			reqBlockNum, _ := req.Params[0].(string)
			isFull, _ := req.Params[1].(bool)
			blockHeader := map[string]interface{}{
				"number":           "0x1000",
				"hash":             "0x0000000000000000000000000000000000000000000000000000000000000001",
				"parentHash":       "0x0000000000000000000000000000000000000000000000000000000000000002",
				"sha3Uncles":       "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
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
				"transactionsRoot": "0x0000000000000000000000000000000000000000000000000000000000000001",
				"logsBloom":        "0x" + strings.Repeat("00", 256),
				"transactions":     []interface{}{},
			}

			if reqBlockNum == "0x1000" || reqBlockNum == "" {
				if isFull {
					blockHeader["transactions"] = []map[string]interface{}{
						{
							"from":        fromAddress,
							"to":          targetAddress,
							"hash":        txHash,
							"value":       "0xde0b6b3a7640000",
							"gas":         "0x5208",
							"gasPrice":    "0x4a817c800",
							"nonce":       "0x1",
							"blockNumber": "0x1000",
							"input":       "0x",
							"v":           "0x" + sigV.Text(16),
							"r":           "0x" + sigR.Text(16),
							"s":           "0x" + sigS.Text(16),
							"type":        "0x0",
						},
					}
				}
			} else {
				blockHeader["transactionsRoot"] = "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421"
			}
			result = blockHeader
		default:
			result = "0x0"
		}

		resp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  result,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	txs, _, err := FetchTransactions(targetAddress, []string{server.URL}, 4)
	if err != nil {
		t.Fatalf("FetchTransactions returned error: %v", err)
	}

	if len(txs) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(txs))
	}

	tx := txs[0]
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
}
