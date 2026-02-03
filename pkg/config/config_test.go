package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfig_Malformed(t *testing.T) {
	reader := strings.NewReader(`{ "addresses": [`)
	_, _, _, _, err := LoadConfig(reader)
	if err == nil {
		t.Error("Expected error loading malformed config, got nil")
	}
}

func TestSaveConfig(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "test_save_config_*.json")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmpfile.Name()
	_ = tmpfile.Close()
	defer func() { _ = os.Remove(tmpPath) }()

	addresses := []AddressConfig{{Address: "0x123", Name: "Test"}}
	chains := []ChainConfig{{
		Name:    "Ethereum",
		RPCURLs: []string{"http://localhost:8545"},
	}}
	globalCfg := GlobalConfig{PrivacyTimeoutSeconds: 120}

	err = SaveConfig(addresses, chains, 0, globalCfg, tmpPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	loadedAddrs, loadedChains, loadedIdx, loadedGlobal, err := LoadConfigFromFile(tmpPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if len(loadedAddrs) != 1 || loadedAddrs[0].Address != "0x123" {
		t.Errorf("Address mismatch")
	}
	if len(loadedChains) != 1 || loadedChains[0].Name != "Ethereum" {
		t.Errorf("Chain mismatch")
	}
	if loadedIdx != 0 {
		t.Errorf("Selected index mismatch")
	}
	if loadedGlobal.PrivacyTimeoutSeconds != 120 {
		t.Errorf("Global config mismatch")
	}
}

func TestLoadConfig_TableDriven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		jsonContent string
		expectError bool
		validate    func(*testing.T, []AddressConfig, []ChainConfig, GlobalConfig)
	}{
		{
			name: "Valid Modern Config",
			jsonContent: `{
				"addresses": [{"address": "0x123", "name": "Main"}],
				"chains": [{"name": "Eth", "rpc_urls": ["http://eth"]}],
				"selected_chain": "Eth",
				"privacy_timeout_seconds": 100
			}`,
			expectError: false,
			validate: func(t *testing.T, addrs []AddressConfig, chains []ChainConfig, g GlobalConfig) {
				if len(addrs) != 1 || addrs[0].Address != "0x123" {
					t.Errorf("Address mismatch")
				}
				if len(chains) != 1 || chains[0].Name != "Eth" {
					t.Errorf("Chain mismatch")
				}
				if g.PrivacyTimeoutSeconds != 100 {
					t.Errorf("Global config mismatch")
				}
			},
		},
		{
			name: "Legacy Addresses (String Array)",
			jsonContent: `{
				"addresses": ["0x123", "0x456"],
				"chains": [{"name": "Eth", "rpc_urls": ["http://eth"]}]
			}`,
			expectError: false,
			validate: func(t *testing.T, addrs []AddressConfig, chains []ChainConfig, g GlobalConfig) {
				if len(addrs) != 2 {
					t.Errorf("Expected 2 addresses, got %d", len(addrs))
				}
				if addrs[0].Address != "0x123" || addrs[1].Address != "0x456" {
					t.Errorf("Address content mismatch")
				}
			},
		},
		{
			name: "Legacy Chains (Root RPC URLs)",
			jsonContent: `{
				"addresses": [{"address": "0x123"}],
				"rpc_urls": ["http://legacy-rpc"]
			}`,
			expectError: false,
			validate: func(t *testing.T, addrs []AddressConfig, chains []ChainConfig, g GlobalConfig) {
				if len(chains) != 1 {
					t.Fatalf("Expected 1 chain from legacy migration, got %d", len(chains))
				}
				if chains[0].Name != "Ethereum" {
					t.Errorf("Expected default name 'Ethereum', got %s", chains[0].Name)
				}
				if len(chains[0].RPCURLs) != 1 || chains[0].RPCURLs[0] != "http://legacy-rpc" {
					t.Errorf("RPC URL mismatch")
				}
			},
		},
		{
			name:        "Malformed JSON",
			jsonContent: `{ "addresses": [ unclosed_array`,
			expectError: true,
			validate:    nil,
		},
		{
			name: "Partial Config (Defaults)",
			jsonContent: `{
				"addresses": [{"address": "0x123"}],
				"chains": [{"name": "Eth", "rpc_urls": ["http://eth"]}]
			}`,
			expectError: false,
			validate: func(t *testing.T, addrs []AddressConfig, chains []ChainConfig, g GlobalConfig) {
				if g.PrivacyTimeoutSeconds != 60 {
					t.Errorf("Expected default privacy timeout 60, got %d", g.PrivacyTimeoutSeconds)
				}
				if g.FiatDecimals != 2 {
					t.Errorf("Expected default fiat decimals 2, got %d", g.FiatDecimals)
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reader := strings.NewReader(tt.jsonContent)
			addrs, chains, _, gCfg, err := LoadConfig(reader)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tt.validate != nil {
					tt.validate(t, addrs, chains, gCfg)
				}
			}
		})
	}
}

func TestSaveConfig_PermissionError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "readonly_test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := os.Chmod(tmpDir, 0500); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(tmpDir, 0700) }()

	configPath := filepath.Join(tmpDir, "config.json")

	addresses := []AddressConfig{{Address: "0x123"}}
	chains := []ChainConfig{{Name: "Eth", RPCURLs: []string{"http://eth"}}}
	globalCfg := GlobalConfig{}

	err = SaveConfig(addresses, chains, 0, globalCfg, configPath)
	if err == nil {
		t.Error("Expected permission error, got nil")
	}
}
