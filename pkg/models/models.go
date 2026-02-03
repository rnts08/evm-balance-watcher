package models

import (
	"math/big"
	"time"
)

// Transaction holds basic transaction details.
type Transaction struct {
	Hash        string
	From        string
	To          string
	Value       string
	BlockNumber uint64
	GasLimit    uint64
	GasPrice    string
	Nonce       uint64
}

// Account holds the data for a single monitored address.
type Account struct {
	Address       string
	Name          string
	Balances      map[string]*big.Float            // Key: Chain Name
	TokenBalances map[string]map[string]*big.Float // Key: Chain Name -> Token Symbol
	Balances24h   map[string]*big.Float            // Key: Chain Name
	Errors        map[string]error                 // Key: Chain Name
	Transactions  []Transaction
}

// AccountChainData holds fetched data for an account on a specific chain.
type AccountChainData struct {
	Address       string
	Balance       *big.Float
	Balance24h    *big.Float
	TokenBalances map[string]*big.Float
}

// ChainData contains the result of a bulk fetch for a chain.
type ChainData struct {
	ChainName  string
	Results    []AccountChainData
	FailedRPCs []string
	Err        error
}

// PriceData contains the current ETH price in USD.
type PriceData struct {
	CoinID string
	Price  float64
	Err    error
}

// GasPriceData contains the current gas price.
type GasPriceData struct {
	Price      *big.Int
	FailedRPCs []string
	Err        error
}

// GasPricePoint holds a timestamped gas price value.
type GasPricePoint struct {
	Timestamp time.Time
	Value     float64
}

// RPCLatencyData contains the result of a latency check.
type RPCLatencyData struct {
	RPCURL  string
	Latency time.Duration
	Err     error
}

// TokenMetadata contains the result of a token metadata fetch.
type TokenMetadata struct {
	Symbol   string
	Decimals int
	Err      error
}

// ChainResult holds test results for a specific chain.
type ChainResult struct {
	Name            string      `json:"name"`
	Symbol          string      `json:"symbol"`
	ConfigChainID   int64       `json:"config_chain_id"`
	RPCs            []RPCResult `json:"rpcs"`
	Inconsistent    bool        `json:"inconsistent"`
	ChainIDUpdated  bool        `json:"chain_id_updated"`
	ObservedChainID int64       `json:"observed_chain_id,omitempty"`
}

// RPCResult holds test results for a specific RPC URL.
type RPCResult struct {
	URL     string `json:"url"`
	Status  string `json:"status"` // "ok" or "error"
	ChainID int64  `json:"chain_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// TestReport holds the results of the configuration test.
type TestReport struct {
	ConfigPath         string        `json:"config_path"`
	ValidStructure     bool          `json:"valid_structure"`
	StructureErrors    []string      `json:"structure_errors,omitempty"`
	AddressCount       int           `json:"address_count"`
	ChainCount         int           `json:"chain_count"`
	Chains             []ChainResult `json:"chains,omitempty"`
	InconsistentChains []string      `json:"inconsistent_chains,omitempty"`
	ConfigUpdated      bool          `json:"config_updated"`
	SaveError          string        `json:"save_error,omitempty"`
	DryRun             bool          `json:"dry_run"`
}
