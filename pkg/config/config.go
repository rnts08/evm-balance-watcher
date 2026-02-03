package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const ConfigFileName = ".evmbal.json"

// TokenConfig holds configuration for an ERC-20 token.
type TokenConfig struct {
	Symbol      string `json:"symbol"`
	Address     string `json:"address"`
	Decimals    int    `json:"decimals"`
	CoinGeckoID string `json:"coingecko_id"`
}

// AddressConfig holds configuration for a monitored address.
type AddressConfig struct {
	Address string `json:"address"`
	Name    string `json:"name,omitempty"`
}

// ChainConfig holds configuration for a specific EVM chain.
type ChainConfig struct {
	Name        string        `json:"name"`
	RPCURLs     []string      `json:"rpc_urls"`
	Symbol      string        `json:"symbol"`
	CoinGeckoID string        `json:"coingecko_id"`
	ChainID     int64         `json:"chain_id,omitempty"`
	ExplorerURL string        `json:"explorer_url,omitempty"`
	Tokens      []TokenConfig `json:"tokens"`
}

// GlobalConfig holds application-wide settings.
type GlobalConfig struct {
	PrivacyTimeoutSeconds    int  `json:"privacy_timeout_seconds"`
	FiatDecimals             int  `json:"fiat_decimals"`
	TokenDecimals            int  `json:"token_decimals"`
	AutoCycleEnabled         bool `json:"auto_cycle_enabled"`
	AutoCycleIntervalSeconds int  `json:"auto_cycle_interval_seconds"`
}

func GetConfigPath(customPath string) (string, error) {
	if customPath != "" {
		return customPath, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigFileName), nil
}

func LoadConfigFromFile(path string) ([]AddressConfig, []ChainConfig, int, GlobalConfig, error) {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return []AddressConfig{}, nil, 0, GlobalConfig{PrivacyTimeoutSeconds: 60, FiatDecimals: 2, TokenDecimals: 2}, nil
	}
	if err != nil {
		return nil, nil, 0, GlobalConfig{}, err
	}
	defer func() { _ = f.Close() }()
	return LoadConfig(f)
}

func LoadConfig(r io.Reader) ([]AddressConfig, []ChainConfig, int, GlobalConfig, error) {
	var cfg struct {
		Addresses                json.RawMessage `json:"addresses"`
		RPCURLs                  []string        `json:"rpc_urls"` // Legacy
		Chains                   []ChainConfig   `json:"chains"`
		SelectedChain            string          `json:"selected_chain"`
		PrivacyTimeoutSeconds    *int            `json:"privacy_timeout_seconds"`
		FiatDecimals             *int            `json:"fiat_decimals"`
		TokenDecimals            *int            `json:"token_decimals"`
		AutoCycleEnabled         *bool           `json:"auto_cycle_enabled"`
		AutoCycleIntervalSeconds *int            `json:"auto_cycle_interval_seconds"`
	}
	if err := json.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, nil, 0, GlobalConfig{}, err
	}

	var addresses []AddressConfig
	// Try unmarshal as []AddressConfig
	if err := json.Unmarshal(cfg.Addresses, &addresses); err != nil {
		addresses = nil
		// Try unmarshal as []string (legacy)
		var strAddrs []string
		if err2 := json.Unmarshal(cfg.Addresses, &strAddrs); err2 == nil {
			for _, a := range strAddrs {
				addresses = append(addresses, AddressConfig{Address: a})
			}
		}
	}

	// Migration for legacy config
	if len(cfg.Chains) == 0 && len(cfg.RPCURLs) > 0 {
		cfg.Chains = []ChainConfig{{
			Name:        "Ethereum",
			RPCURLs:     cfg.RPCURLs,
			Symbol:      "ETH",
			CoinGeckoID: "ethereum",
			ExplorerURL: "https://etherscan.io",
		}}
		cfg.SelectedChain = "Ethereum"
	}

	selectedIdx := 0
	for i, c := range cfg.Chains {
		if c.Name == cfg.SelectedChain {
			selectedIdx = i
			break
		}
	}

	globalCfg := GlobalConfig{
		PrivacyTimeoutSeconds:    60,
		FiatDecimals:             2,
		TokenDecimals:            2,
		AutoCycleEnabled:         false,
		AutoCycleIntervalSeconds: 15,
	}
	if cfg.PrivacyTimeoutSeconds != nil {
		globalCfg.PrivacyTimeoutSeconds = *cfg.PrivacyTimeoutSeconds
	}
	if cfg.FiatDecimals != nil {
		globalCfg.FiatDecimals = *cfg.FiatDecimals
	}
	if cfg.TokenDecimals != nil {
		globalCfg.TokenDecimals = *cfg.TokenDecimals
	}
	if cfg.AutoCycleEnabled != nil {
		globalCfg.AutoCycleEnabled = *cfg.AutoCycleEnabled
	}
	if cfg.AutoCycleIntervalSeconds != nil {
		globalCfg.AutoCycleIntervalSeconds = *cfg.AutoCycleIntervalSeconds
	}

	return addresses, cfg.Chains, selectedIdx, globalCfg, nil
}

func SaveConfig(addresses []AddressConfig, chains []ChainConfig, selectedIdx int, globalCfg GlobalConfig, path string) error {
	// Validation: Ensure we have at least one chain
	if len(chains) == 0 {
		return fmt.Errorf("validation failed: configuration must have at least one chain")
	}

	// Validation: Ensure chains have names and RPCs
	for i, c := range chains {
		if strings.TrimSpace(c.Name) == "" {
			return fmt.Errorf("validation failed: chain at index %d has no name", i)
		}
		if len(c.RPCURLs) == 0 {
			return fmt.Errorf("validation failed: chain %s has no RPC URLs", c.Name)
		}
	}

	selectedName := ""
	if selectedIdx >= 0 && selectedIdx < len(chains) {
		selectedName = chains[selectedIdx].Name
	}
	cfg := struct {
		Addresses                []AddressConfig `json:"addresses"`
		Chains                   []ChainConfig   `json:"chains"`
		SelectedChain            string          `json:"selected_chain"`
		PrivacyTimeoutSeconds    int             `json:"privacy_timeout_seconds"`
		FiatDecimals             int             `json:"fiat_decimals"`
		TokenDecimals            int             `json:"token_decimals"`
		AutoCycleEnabled         bool            `json:"auto_cycle_enabled"`
		AutoCycleIntervalSeconds int             `json:"auto_cycle_interval_seconds"`
	}{
		Addresses:                addresses,
		Chains:                   chains,
		SelectedChain:            selectedName,
		PrivacyTimeoutSeconds:    globalCfg.PrivacyTimeoutSeconds,
		FiatDecimals:             globalCfg.FiatDecimals,
		TokenDecimals:            globalCfg.TokenDecimals,
		AutoCycleEnabled:         globalCfg.AutoCycleEnabled,
		AutoCycleIntervalSeconds: globalCfg.AutoCycleIntervalSeconds,
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("validation failed: encoded configuration is empty")
	}

	// Create a backup of the existing file
	if _, err := os.Stat(path); err == nil {
		backupPath := fmt.Sprintf("%s.%s.bak", path, time.Now().Format("20060102-150405"))
		input, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read existing config for backup: %w", err)
		}
		if err := os.WriteFile(backupPath, input, 0644); err != nil {
			return fmt.Errorf("failed to write backup config: %w", err)
		}
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func RestoreLastBackup(configPath string) error {
	matches, err := filepath.Glob(configPath + ".*.bak")
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no backup files found")
	}
	sort.Strings(matches)
	lastBackup := matches[len(matches)-1]

	data, err := os.ReadFile(lastBackup)
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0644)
}
