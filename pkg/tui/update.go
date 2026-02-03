package tui

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"evmbal/pkg/models"
	"evmbal/pkg/watcher"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	activeChain := m.chains[m.activeChainIdx]

	switch msg := msg.(type) {

	// Handle Token Metadata Result (still separate as it's a one-off UI action)
	case models.TokenMetadata:
		if m.addingToken {
			if msg.Err == nil {
				if m.tokenInputs[0].Value() == "" && msg.Symbol != "" {
					m.tokenInputs[0].SetValue(msg.Symbol)
				}
				if m.tokenInputs[2].Value() == "" && msg.Decimals != 0 {
					m.tokenInputs[2].SetValue(strconv.Itoa(msg.Decimals))
				}
				m.statusMessage = "Token metadata fetched!"
			}
			cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
				return clearStatusMsg{}
			}))
		}

	case watcher.Event:
		// Re-subscribe to next event
		cmds = append(cmds, listenForWatcher(m.watcher.Subscribe()))

		switch msg.Type {
		case watcher.EventPriceUpdated:
			if data, ok := msg.Data.(models.PriceData); ok {
				m.prices[data.CoinID] = data.Price
			}
		case watcher.EventChainDataUpdated:
			if data, ok := msg.Data.(models.ChainData); ok {
				m.loading = false
				for _, res := range data.Results {
					for _, acc := range m.accounts {
						if strings.EqualFold(acc.Address, res.Address) {
							if acc.Balances == nil {
								acc.Balances = make(map[string]*big.Float)
							}
							if acc.TokenBalances == nil {
								acc.TokenBalances = make(map[string]map[string]*big.Float)
							}
							if acc.Balances24h == nil {
								acc.Balances24h = make(map[string]*big.Float)
							}
							if acc.Errors == nil {
								acc.Errors = make(map[string]error)
							}
							acc.Balances[data.ChainName] = res.Balance
							acc.Balances24h[data.ChainName] = res.Balance24h
							if acc.TokenBalances[data.ChainName] == nil {
								acc.TokenBalances[data.ChainName] = make(map[string]*big.Float)
							}
							for sym, bal := range res.TokenBalances {
								acc.TokenBalances[data.ChainName][sym] = bal
							}
							delete(acc.Errors, data.ChainName)
							break
						}
					}
				}
			}
		case watcher.EventGasPriceUpdated:
			if data, ok := msg.Data.(models.GasPriceData); ok {
				if m.gasPrice != nil {
					m.gasTrend = data.Price.Cmp(m.gasPrice)
				}
				m.gasPrice = data.Price
				gwei := new(big.Float).Quo(new(big.Float).SetInt(data.Price), big.NewFloat(1e9))
				val, _ := gwei.Float64()
				m.gasPriceHistory = append(m.gasPriceHistory, models.GasPricePoint{Timestamp: time.Now(), Value: val})
				if len(m.gasPriceHistory) > 2880 {
					m.gasPriceHistory = m.gasPriceHistory[len(m.gasPriceHistory)-2880:]
				}
			}
		case watcher.EventTransactionsUpdated:
			if data, ok := msg.Data.(map[string]interface{}); ok {
				addr, _ := data["address"].(string)
				txs, _ := data["txs"].([]models.Transaction)
				for _, acc := range m.accounts {
					if acc.Address == addr {
						acc.Transactions = txs
						break
					}
				}
			}
		}

		m.lastUpdate = time.Now()
		if m.showDetail {
			m.updateDetailViewport()
		}

	case models.RPCLatencyData:
		if m.rpcLatencyHistory == nil {
			m.rpcLatencyHistory = make(map[string][]time.Duration)
		}
		val := msg.Latency
		if msg.Err != nil {
			m.rpcLatencies[msg.RPCURL] = -1
			val = -1
		} else {
			m.rpcLatencies[msg.RPCURL] = msg.Latency
		}
		hist := m.rpcLatencyHistory[msg.RPCURL]
		hist = append(hist, val)
		if len(hist) > 15 {
			hist = hist[len(hist)-15:]
		}
		m.rpcLatencyHistory[msg.RPCURL] = hist

	case privacyTimeoutMsg:
		if m.config.PrivacyTimeoutSeconds <= 0 {
			break
		}
		timeoutDuration := time.Duration(m.config.PrivacyTimeoutSeconds) * time.Second
		if !m.privacyMode {
			if time.Since(m.lastInteraction) >= timeoutDuration {
				m.privacyMode = true
				m.statusMessage = "Privacy Mode enabled due to inactivity"
				cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
					return clearStatusMsg{}
				}))
			} else {
				remaining := timeoutDuration - time.Since(m.lastInteraction)
				cmds = append(cmds, tea.Tick(remaining, func(t time.Time) tea.Msg {
					return privacyTimeoutMsg{}
				}))
			}
		}

	case autoCycleMsg:
		if m.config.AutoCycleEnabled && m.config.AutoCycleIntervalSeconds > 0 {
			if time.Since(m.lastInteraction) < 5*time.Second {
				nextCheck := 1 * time.Second
				m.nextAutoCycleTime = time.Now().Add(nextCheck)
				cmds = append(cmds, tea.Tick(nextCheck, func(t time.Time) tea.Msg {
					return autoCycleMsg{}
				}))
			} else {
				if len(m.accounts) > 1 {
					m.activeIdx = (m.activeIdx + 1) % len(m.accounts)
				}
				interval := time.Duration(m.config.AutoCycleIntervalSeconds) * time.Second
				m.nextAutoCycleTime = time.Now().Add(interval)
				cmds = append(cmds, tea.Tick(interval, func(t time.Time) tea.Msg {
					return autoCycleMsg{}
				}))
			}
		}

	case tea.KeyMsg:
		m.lastInteraction = time.Now()
		isInputMode := m.editingAddress || m.addingToken || m.addingChain || m.adding || m.exportingConfig || m.editingGlobalConfig
		if !isInputMode && msg.String() == "?" {
			m.showHelp = !m.showHelp
			return m, nil
		}
		if m.showHelp {
			if msg.String() == "q" || msg.String() == "esc" || msg.String() == "?" {
				m.showHelp = false
			}
			return m, nil
		}

		if msg.String() == "P" {
			m.privacyMode = !m.privacyMode
			if !m.privacyMode && m.config.PrivacyTimeoutSeconds > 0 {
				cmds = append(cmds, tea.Tick(time.Duration(m.config.PrivacyTimeoutSeconds)*time.Second, func(t time.Time) tea.Msg {
					return privacyTimeoutMsg{}
				}))
			}
		}

		if msg.String() == "A" {
			m.config.AutoCycleEnabled = !m.config.AutoCycleEnabled
			status := "disabled"
			if m.config.AutoCycleEnabled {
				status = "enabled"
				if m.config.AutoCycleIntervalSeconds > 0 {
					interval := time.Duration(m.config.AutoCycleIntervalSeconds) * time.Second
					m.nextAutoCycleTime = time.Now().Add(interval)
					cmds = append(cmds, tea.Tick(interval, func(t time.Time) tea.Msg {
						return autoCycleMsg{}
					}))
				}
			}
			m.statusMessage = fmt.Sprintf("Auto-cycle %s", status)
			cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
				return clearStatusMsg{}
			}))
		}

		if m.showTxDetail {
			switch msg.String() {
			case "q", "esc", "backspace":
				m.showTxDetail = false
				return m, nil
			case "o":
				if activeChain.ExplorerURL == "" {
					m.statusMessage = "Explorer URL not configured for this chain"
					cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
						return clearStatusMsg{}
					}))
					return m, tea.Batch(cmds...)
				}
				acc := m.accounts[m.activeIdx]
				if len(acc.Transactions) > m.txListIdx {
					tx := acc.Transactions[m.txListIdx]
					url := fmt.Sprintf("%s/tx/%s", strings.TrimRight(activeChain.ExplorerURL, "/"), tx.Hash)
					if err := openBrowser(url); err != nil {
						m.statusMessage = fmt.Sprintf("Failed to open browser: %v", err)
					} else {
						m.statusMessage = "Opened in browser"
					}
				}
				cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
					return clearStatusMsg{}
				}))
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}

		if m.showTxList {
			switch msg.String() {
			case "q", "esc":
				m.showTxList = false
				return m, nil
			case "i":
				m.txFilter = "in"
				m.txListIdx = 0
				return m, nil
			case "o":
				m.txFilter = "out"
				m.txListIdx = 0
				return m, nil
			case "a":
				m.txFilter = "all"
				m.txListIdx = 0
				return m, nil
			case "up", "k":
				if m.txListIdx > 0 {
					m.txListIdx--
				}
			case "down", "j":
				txs := m.getFilteredTransactions(m.accounts[m.activeIdx])
				if m.txListIdx < len(txs)-1 {
					m.txListIdx++
				}
			case "enter":
				txs := m.getFilteredTransactions(m.accounts[m.activeIdx])
				if len(txs) > 0 {
					m.showTxDetail = true
				}
			}
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			if m.showSummary {
				m.showSummary = false
				return m, nil
			}
			if m.showNetworkStatus {
				m.showNetworkStatus = false
				return m, nil
			}
			if m.showGasTracker {
				m.showGasTracker = false
				return m, nil
			}
			if m.editingAddress || m.adding || m.addingChain || m.addingToken || m.editingGlobalConfig || m.exportingConfig || m.restoringBackup {
				return m, nil
			}
			return m, tea.Quit
		case "G":
			m.showGasTracker = true
			return m, nil
		case "r":
			m.loading = true
			// Manual refresh: in the new world, we tell the watcher to fetch now
			// For now, it's automatic anyway.
			m.statusMessage = "Refreshing data..."
			cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
				return clearStatusMsg{}
			}))

		case "enter":
			if len(m.accounts) > 0 {
				m.showDetail = true
				m.updateDetailViewport()
				m.viewport.YOffset = 0
			}

		case "c":
			if len(m.accounts) > 0 {
				err := clipboard.WriteAll(m.accounts[m.activeIdx].Address)
				if err != nil {
					m.statusMessage = "Failed to copy to clipboard"
				} else {
					if m.privacyMode {
						m.statusMessage = "Full address copied (Privacy Mode active)!"
					} else {
						m.statusMessage = "Full address copied to clipboard!"
					}
				}
				cmds = append(cmds, tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
					return clearStatusMsg{}
				}))
			}

		case "tab", "right", "l":
			if len(m.accounts) > 0 {
				m.activeIdx++
				if m.activeIdx >= len(m.accounts) {
					m.activeIdx = 0
				}
			}
		case "shift+tab", "left", "h":
			if len(m.accounts) > 0 {
				m.activeIdx--
				if m.activeIdx < 0 {
					m.activeIdx = len(m.accounts) - 1
				}
			}
		}

	case uiTickMsg:
		cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg { return uiTickMsg(t) }))

	case clearStatusMsg:
		m.statusMessage = ""
	}

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
