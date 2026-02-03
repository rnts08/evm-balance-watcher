package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strings"

	"evmbal/pkg/config"
	"evmbal/pkg/models"
	"evmbal/pkg/server"
	"evmbal/pkg/tui"
	"evmbal/pkg/watcher"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Version should be set during build
var Version = "dev"

func main() {
	testFlag := flag.Bool("t", false, "Test configuration and exit")
	testLongFlag := flag.Bool("test", false, "Test configuration and exit")
	jsonFlag := flag.Bool("json", false, "Output test results as JSON")
	dryRunFlag := flag.Bool("dry-run", false, "Perform a trial run with no changes made")
	configFlag := flag.String("config", "", "Path to configuration file")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	serverFlag := flag.Bool("server", false, "Run in headless server mode")
	portFlag := flag.Int("port", 8080, "Port for API server")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("evmbal version %s\n", Version)
		os.Exit(0)
	}

	cfgInput := *configFlag
	if cfgInput == "" && len(flag.Args()) > 0 {
		cfgInput = flag.Args()[0]
	}
	path, err := config.GetConfigPath(cfgInput)
	if err != nil {
		fmt.Printf("Error determining config path: %v\n", err)
		os.Exit(1)
	}

	savedAddrs, savedChains, activeChainIdx, savedGlobalCfg, err := config.LoadConfigFromFile(path)
	if err != nil {
		fmt.Printf("Error loading config from %s: %v\n", path, err)
		os.Exit(1)
	}

	if *testFlag || *testLongFlag {
		var report models.TestReport
		report.ConfigPath = path
		report.ValidStructure = true
		report.DryRun = *dryRunFlag

		if !*jsonFlag {
			fmt.Printf("Testing configuration at: %s\n", path)
		}

		if len(savedChains) == 0 {
			report.ValidStructure = false
			report.StructureErrors = append(report.StructureErrors, "No Chains found in configuration.")
			if !*jsonFlag {
				fmt.Println("No Chains found in configuration.")
			} else {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				_ = enc.Encode(report)
			}
			os.Exit(1)
		}

		for i, chain := range savedChains {
			if strings.TrimSpace(chain.Name) == "" {
				msg := fmt.Sprintf("Chain at index %d has no name.", i)
				report.StructureErrors = append(report.StructureErrors, msg)
				report.ValidStructure = false
				if !*jsonFlag {
					fmt.Printf("Error: %s\n", msg)
				}
			}
			if len(chain.RPCURLs) == 0 {
				msg := fmt.Sprintf("Chain '%s' has no RPC URLs.", chain.Name)
				report.StructureErrors = append(report.StructureErrors, msg)
				report.ValidStructure = false
				if !*jsonFlag {
					fmt.Printf("Error: %s\n", msg)
				}
			}
		}

		if !report.ValidStructure {
			if *jsonFlag {
				enc := json.NewEncoder(os.Stdout)
				enc.SetIndent("", "  ")
				_ = enc.Encode(report)
			}
			os.Exit(1)
		}

		report.AddressCount = len(savedAddrs)
		report.ChainCount = len(savedChains)

		if !*jsonFlag {
			fmt.Printf("Found %d addresses and %d Chains.\n", len(savedAddrs), len(savedChains))
		}

		var inconsistentChains []string
		configUpdated := false
		for i := range savedChains {
			chain := &savedChains[i]
			cResult := models.ChainResult{
				Name:          chain.Name,
				Symbol:        chain.Symbol,
				ConfigChainID: chain.ChainID,
			}

			if !*jsonFlag {
				fmt.Printf("Testing Chain: %s (%s)\n", chain.Name, chain.Symbol)
			}
			var observedChainID *big.Int
			chainInconsistent := false
			for _, rpc := range chain.RPCURLs {
				rResult := models.RPCResult{URL: rpc}
				if !*jsonFlag {
					fmt.Printf("  RPC: %s ... ", rpc)
				}
				client, err := ethclient.Dial(rpc)
				if err != nil {
					rResult.Status = "error"
					rResult.Error = err.Error()
					if !*jsonFlag {
						fmt.Printf("Failed: %v\n", err)
					}
					cResult.RPCs = append(cResult.RPCs, rResult)
					continue
				}
				id, err := client.ChainID(context.Background())
				if err != nil {
					rResult.Status = "error"
					rResult.Error = fmt.Sprintf("Failed to get ChainID: %v", err)
					if !*jsonFlag {
						fmt.Printf("Failed to get ChainID: %v\n", err)
					}
				} else {
					rResult.Status = "ok"
					rResult.ChainID = id.Int64()
					if !*jsonFlag {
						fmt.Printf("OK (ChainID: %s)", id.String())
					}
					if observedChainID == nil {
						observedChainID = id
						cResult.ObservedChainID = id.Int64()
					} else if observedChainID.Cmp(id) != 0 {
						if !*jsonFlag {
							fmt.Printf(" - WARNING: ChainID mismatch with previous RPC (%s)", observedChainID.String())
						}
						chainInconsistent = true
					}

					if chain.ChainID != 0 {
						if id.Cmp(big.NewInt(chain.ChainID)) != 0 {
							rResult.Error = fmt.Sprintf("Mismatch! Expected %d", chain.ChainID)
							if !*jsonFlag {
								fmt.Printf(" - MISMATCH! Expected %d", chain.ChainID)
							}
						} else {
							if !*jsonFlag {
								fmt.Printf(" - Verified")
							}
						}
					} else {
						chain.ChainID = id.Int64()
						configUpdated = true
						cResult.ChainIDUpdated = true
						if !*jsonFlag {
							fmt.Printf(" - UPDATED CONFIG")
							if *dryRunFlag {
								fmt.Printf(" (DRY RUN)")
							}
						}
					}
					if !*jsonFlag {
						fmt.Println()
					}
				}
				client.Close()
				cResult.RPCs = append(cResult.RPCs, rResult)
			}
			if chainInconsistent {
				cResult.Inconsistent = true
				inconsistentChains = append(inconsistentChains, chain.Name)
			}
			report.Chains = append(report.Chains, cResult)
		}

		report.InconsistentChains = inconsistentChains

		if len(inconsistentChains) > 0 {
			if !*jsonFlag {
				fmt.Println("\nWARNING: Inconsistent RPCs detected!")
				fmt.Println("The following chains have RPCs returning conflicting Chain IDs:")
				for _, name := range inconsistentChains {
					fmt.Printf(" - %s\n", name)
				}
			}
		}

		if configUpdated {
			report.ConfigUpdated = true
			if !*jsonFlag {
				fmt.Println("\nUpdating configuration with fetched Chain IDs...")
			}
			if *dryRunFlag {
				if !*jsonFlag {
					fmt.Println("Dry run enabled: Configuration NOT saved.")
				}
			} else {
				if err := config.SaveConfig(savedAddrs, savedChains, activeChainIdx, savedGlobalCfg, path); err != nil {
					report.SaveError = err.Error()
					if !*jsonFlag {
						fmt.Printf("Failed to save config: %v\n", err)
					}
				} else {
					if !*jsonFlag {
						fmt.Println("Configuration saved successfully.")
					}
				}
			}
		}

		if *jsonFlag {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(report)
		}
		os.Exit(0)
	}

	if len(savedChains) == 0 {
		fmt.Println("Error: No Chains found in configuration.")
		fmt.Printf("Please create a config file at %s with 'chains'.\n", path)
		os.Exit(1)
	}

	w := watcher.NewWatcher(savedAddrs, savedChains, savedGlobalCfg, path)
	go w.Start(context.Background())

	srv := server.NewServer(w)
	go func() {
		if err := srv.Start(*portFlag); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	if *serverFlag {
		fmt.Printf("Running in server mode on port %d...\n", *portFlag)
		select {} // Keep alive
	}

	tui.Start(w, savedAddrs, savedChains, activeChainIdx, savedGlobalCfg, path, Version)
}
