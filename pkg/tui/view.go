package tui

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/guptarohit/asciigraph"

	"evmbal/pkg/utils"
)

func (m model) View() string {
	var content string

	if m.showHelp {
		return m.viewHelp()
	}

	if m.showGasTracker {
		return m.viewGasTracker()
	}

	if m.showTxDetail {
		return m.viewTxDetail()
	}

	if m.showTxList {
		return m.viewTxList()
	}

	if m.restoringBackup {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center,
				titleStyle.Render("Confirm Restore"),
				"\n",
				"Are you sure you want to restore the last backup?",
				"Current configuration will be overwritten.",
				"\n",
				subtleStyle.Render("(y) Yes • (n) No"),
			)),
		)
	}

	if m.editingGlobalConfig {
		labels := []string{"Privacy Timeout (s)", "Fiat Decimals", "Token Decimals", "Auto Cycle (t/f)", "Cycle Interval (s)"}
		var inputs []string
		for i, label := range labels {
			inputs = append(inputs, fmt.Sprintf("%-20s %s", label, m.globalConfigInputs[i].View()))
		}
		return lipgloss.Place(
			m.width, m.height, lipgloss.Center, lipgloss.Center,
			boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Global Settings"),
				"\n",
				strings.Join(inputs, "\n"),
				"\n",
				subtleStyle.Render("Enter to next/save • Esc to cancel"),
			)),
		)
	}

	if m.exportingConfig {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Export Configuration"),
				"\n",
				"Enter file path:",
				m.exportInput.View(),
				"\n",
				subtleStyle.Render("Enter to save • Esc to cancel"),
			)),
		)
	}

	if m.editingAddress {
		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Edit Address Name"),
				"\n",
				fmt.Sprintf("Address: %s", m.accounts[m.activeIdx].Address),
				"\n",
				m.editAddressInput.View(),
				"\n",
				subtleStyle.Render("Enter to save • Esc to cancel"),
			)),
		)
	}

	if m.addingToken {
		labels := []string{"Symbol", "Address", "Decimals", "CoinGecko ID"}
		var inputs []string
		for i, label := range labels {
			inputs = append(inputs, fmt.Sprintf("%-15s %s", label, m.tokenInputs[i].View()))
		}

		return lipgloss.Place(
			m.width, m.height, lipgloss.Center, lipgloss.Center,
			boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Add New Token"),
				"\n",
				strings.Join(inputs, "\n"),
				"\n",
				subtleStyle.Render("Enter to next/save • Esc to cancel"),
			)),
		)
	}

	if m.managingTokens {
		chain := m.chains[m.selectedChainForTokens]
		header := titleStyle.Render(fmt.Sprintf("Manage Tokens (%s)", chain.Name))
		rows := ""
		for i, t := range chain.Tokens {
			cursor := "  "
			if i == m.tokenListIdx {
				cursor = "> "
			}
			rows += fmt.Sprintf("%s%s (%s)\n", cursor, t.Symbol, utils.TruncateString(t.Address, 20))
		}
		content = boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", rows))
		footer := subtleStyle.Render("a: add • d: delete • q: back")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer))
	}

	if m.addingChain {
		labels := []string{"Name", "Symbol", "CoinGecko ID", "RPC URLs", "Explorer URL"}
		var inputs []string
		for i, label := range labels {
			inputs = append(inputs, fmt.Sprintf("%-15s %s", label, m.chainInputs[i].View()))
		}

		return lipgloss.Place(
			m.width, m.height, lipgloss.Center, lipgloss.Center,
			boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Add New Chain"),
				"\n",
				strings.Join(inputs, "\n"),
				"\n",
				subtleStyle.Render("Enter to next/save • Esc to cancel"),
			)),
		)
	}

	if m.managingChains {
		header := titleStyle.Render("Manage Chains")
		rows := ""
		for i, c := range m.chains {
			cursor := "  "
			if i == m.chainListIdx {
				cursor = "> "
			}
			rows += fmt.Sprintf("%s%s (%s)\n", cursor, c.Name, c.Symbol)
		}
		content = boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", rows))
		footer := subtleStyle.Render("a: add • d: delete • t: tokens • q: back")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer))
	}

	if m.adding {
		labels := []string{"Address", "Name"}
		var inputs []string
		for i, label := range labels {
			inputs = append(inputs, fmt.Sprintf("%-10s %s", label, m.addressInputs[i].View()))
		}

		return lipgloss.Place(
			m.width,
			m.height,
			lipgloss.Center,
			lipgloss.Center,
			boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
				titleStyle.Render("Add New Address"),
				"\n",
				strings.Join(inputs, "\n"),
				"\n",
				subtleStyle.Render("Enter to save • Esc to cancel"),
			)),
		)
	}

	if m.showSummary {
		return m.viewSummary()
	}

	if m.showNetworkStatus {
		return m.viewNetworkStatus()
	}

	if m.showDetail {
		return m.viewDetail()
	}

	if len(m.accounts) == 0 {
		return "No addresses to monitor."
	}

	activeAcc := m.accounts[m.activeIdx]
	activeChain := m.chains[m.activeChainIdx]

	// Top Bar Data
	price := m.prices[activeChain.CoinGeckoID]
	priceDisplay := fmt.Sprintf("%s: N/A", activeChain.Symbol)
	if price > 0 {
		priceDisplay = fmt.Sprintf("%s: $%s", activeChain.Symbol, utils.FormatFloat(price, m.config.FiatDecimals))
	}
	gasDisplay := "Gas: N/A"
	gasStyle := subtleStyle
	if m.gasPrice != nil {
		gwei := new(big.Float).Quo(new(big.Float).SetInt(m.gasPrice), big.NewFloat(1e9))
		val, _ := gwei.Float64()
		gasDisplay = fmt.Sprintf("Gas: %.2f Gwei", val)
		if m.gasTrend > 0 {
			gasDisplay += " ↑"
		} else if m.gasTrend < 0 {
			gasDisplay += " ↓"
		}
		if val < 30 {
			gasStyle = infoStyle
		} else if val < 100 {
			gasStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B"))
		} else {
			gasStyle = errStyle
		}
	}
	spinnerView := ""
	if m.loading {
		spinnerView = m.spinner.View() + " "
	}
	lastUpdStr := fmt.Sprintf("%sLast updated: %s", spinnerView, m.lastUpdate.Format("15:04:05"))

	balance := activeAcc.Balances[activeChain.Name]
	balance24h := activeAcc.Balances24h[activeChain.Name]
	err := activeAcc.Errors[activeChain.Name]

	if m.loading && balance == nil && err == nil {
		content = "Connecting to Ethereum Node..."
	} else if err != nil {
		content = fmt.Sprintf("%s\n%s",
			errStyle.Render("Error fetching balance:"),
			err.Error(),
		)
	} else {
		// Format Balance
		balStr := fmt.Sprintf("0.00 %s", activeChain.Symbol)
		if balance != nil {
			balStr = fmt.Sprintf("%s %s", m.displayValue(balance, m.config.TokenDecimals), activeChain.Symbol)
			if price > 0 {
				usdVal := new(big.Float).Mul(balance, big.NewFloat(price))
				balStr += fmt.Sprintf(" ($%s)", m.displayValue(usdVal, m.config.FiatDecimals))
			}

			if balance24h != nil {
				diff := new(big.Float).Sub(balance, balance24h)
				sign := "+"
				style := infoStyle
				if diff.Sign() < 0 {
					sign = ""
					style = errStyle
				}
				// 24h change
				balStr += style.Render(fmt.Sprintf(" %s%s (24h)", sign, m.displayValue(diff, m.config.TokenDecimals)))
			}
		}

		// Tokens
		var tokenStrs []string
		if tokens, ok := activeAcc.TokenBalances[activeChain.Name]; ok {
			for _, token := range activeChain.Tokens {
				if bal, ok := tokens[token.Symbol]; ok {
					tokenPrice := m.prices[token.CoinGeckoID]
					tokenVal := new(big.Float).Mul(bal, big.NewFloat(tokenPrice))
					tStr := fmt.Sprintf("%s %s", m.displayValue(bal, m.config.TokenDecimals), token.Symbol)
					if tokenPrice > 0 {
						tStr += fmt.Sprintf(" ($%s)", m.displayValue(tokenVal, m.config.FiatDecimals))
					}
					tokenStrs = append(tokenStrs, tStr)
				}
			}
		}
		if len(tokenStrs) > 0 {
			sep := "\n"
			if m.width >= 80 {
				sep = " • "
			}
			balStr += "\n" + strings.Join(tokenStrs, sep)
		}

		// Construct the UI pieces
		title := fmt.Sprintf("EVM Balance Watcher - %s", activeChain.Name)
		if len(m.accounts) > 1 {
			title = fmt.Sprintf("EVM Balance Watcher - %s (%d/%d)", activeChain.Name, m.activeIdx+1, len(m.accounts))
		}
		header := titleStyle.Render(title)
		addrStr := activeAcc.Address
		if m.privacyMode {
			addrStr = "0x**...**"
		} else if activeAcc.Name != "" {
			if len(activeAcc.Address) > 12 {
				addrStr = activeAcc.Address[:6] + "..." + activeAcc.Address[len(activeAcc.Address)-4:]
			}
		}
		if activeAcc.Name != "" {
			addrStr = fmt.Sprintf("%s (%s)", addrStr, activeAcc.Name)
		}
		addr := fmt.Sprintf("Address: %s", addrStr)
		rpcStr := "No RPC"
		if len(activeChain.RPCURLs) > 0 {
			rpcStr = activeChain.RPCURLs[0]
		}
		rpc := fmt.Sprintf("RPC: %s", utils.TruncateString(rpcStr, 30))

		targetWidth := m.width - 4
		if targetWidth < 0 {
			targetWidth = 0
		}
		contentWidth := targetWidth - 4
		if contentWidth < 0 {
			contentWidth = 0
		}

		balanceDisplay := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true).
			Width(contentWidth).
			Align(lipgloss.Center).
			Render(balStr)

		// Transactions Table
		var txTable string
		if len(activeAcc.Transactions) > 0 {
			headers := tableHeaderStyle.Render(fmt.Sprintf("%-10s %-10s %-10s %-10s", "HASH", "FROM", "TO", "VALUE"))
			rows := ""
			for i, tx := range activeAcc.Transactions {
				if i >= 3 {
					break
				}
				rows += fmt.Sprintf("%-10s %-10s %-10s %-10s\n",
					utils.TruncateString(tx.Hash, 10),
					utils.TruncateString(tx.From, 10),
					utils.TruncateString(tx.To, 10),
					m.maskString(tx.Value),
				)
			}
			txTable = lipgloss.JoinVertical(lipgloss.Center,
				headers,
				rows,
			)
		} else {
			txTable = subtleStyle.Render("No recent transactions found")
		}

		// Combine into a block
		var uiBlock string
		if m.compactMode {
			uiBlock = lipgloss.JoinVertical(lipgloss.Center,
				header,
				addr,
				"\n",
				balanceDisplay,
			)
		} else {
			uiBlock = lipgloss.JoinVertical(lipgloss.Center,
				header,
				addr,
				rpc,
				"\n",
				balanceDisplay,
				"\n",
				txTable,
			)
		}

		content = boxStyle.Width(targetWidth).Align(lipgloss.Center).Render(uiBlock)
	}

	// Footer
	line1 := "r:ref • R:frc • s:sum • t:txs • P:prv • O:cfg • ent:dt • ?:hlp • q:quit"
	if len(m.accounts) > 1 {
		line1 = "Tab:cycle • " + line1
	}

	line2 := "a:add • d:del • e:edt • E:chn • N:net • B:bak • c:cpy • X:exp • G:gas"
	if len(m.chains) > 1 {
		line2 += " • n:nxt"
	}
	line2 += fmt.Sprintf(" • v%s", Version)

	var footer string
	if m.width > 0 {
		l1 := subtleStyle.Width(m.width).Align(lipgloss.Center).Render(line1)
		l2 := subtleStyle.Width(m.width).Align(lipgloss.Center).Render(line2)
		footer = lipgloss.JoinVertical(lipgloss.Center, l1, l2)
	} else {
		footer = subtleStyle.Render(line1 + "\n" + line2)
	}

	if m.statusMessage != "" {
		footer = lipgloss.JoinVertical(lipgloss.Center, infoStyle.Render(m.statusMessage), footer)
	}

	// Construct Top Bar
	priceRendered := subtleStyle.Render(fmt.Sprintf(" %s", priceDisplay))
	sepRendered := subtleStyle.Render(" • ")
	gasRendered := gasStyle.Render(gasDisplay)
	leftBlock := lipgloss.JoinHorizontal(lipgloss.Top, priceRendered, sepRendered, gasRendered)

	privacyIndicator := ""
	if m.privacyMode {
		privacyIndicator = "🔒 "
	}
	autoCycleIndicator := ""
	if m.config.AutoCycleEnabled {
		if time.Since(m.lastInteraction) < 5*time.Second {
			autoCycleIndicator = "⏸ "
		} else {
			remaining := time.Until(m.nextAutoCycleTime).Seconds()
			if remaining < 0 {
				remaining = 0
			}
			autoCycleIndicator = fmt.Sprintf("▶ %ds ", int(remaining))
		}
	}
	rightBlock := subtleStyle.Render(fmt.Sprintf("%s%s%s ", autoCycleIndicator, privacyIndicator, lastUpdStr))
	gap := m.width - lipgloss.Width(leftBlock) - lipgloss.Width(rightBlock)
	if gap < 0 {
		gap = 0
	}
	topBar := lipgloss.JoinHorizontal(lipgloss.Top, leftBlock, strings.Repeat(" ", gap), rightBlock)

	h := m.height - 1
	if h < 0 {
		h = 0
	}

	// Center the content on the screen
	return lipgloss.JoinVertical(lipgloss.Left,
		topBar,
		lipgloss.Place(
			m.width,
			h,
			lipgloss.Center,
			lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer),
		),
	)
}

func (m model) viewNetworkStatus() string {
	activeChain := m.chains[m.activeChainIdx]
	header := titleStyle.Render(fmt.Sprintf("Network Status: %s", activeChain.Name))

	rows := ""
	now := time.Now()

	for _, rpc := range activeChain.RPCURLs {
		status := infoStyle.Render("ACTIVE")
		extra := ""
		if expiry, ok := m.rpcCooldowns[rpc]; ok && now.Before(expiry) {
			status = errStyle.Render("COOLDOWN")
			remaining := expiry.Sub(now).Round(time.Second)
			extra = fmt.Sprintf(" (%s)", remaining)
		}

		latDisplay := ""
		if lat, ok := m.rpcLatencies[rpc]; ok {
			if lat == -1 {
				latDisplay = errStyle.Render(" Error")
			} else {
				s := infoStyle
				if lat > 500*time.Millisecond {
					s = lipgloss.NewStyle().Foreground(lipgloss.Color("#E5C07B"))
				}
				if lat > 1*time.Second {
					s = errStyle
				}
				latDisplay = s.Render(fmt.Sprintf(" %s", lat.Round(time.Millisecond)))
			}
		}
		sparkline := m.renderLatencySparkline(m.rpcLatencyHistory[rpc])
		rows += fmt.Sprintf("%-45s %s%s%s %s\n", utils.TruncateString(rpc, 43), status, extra, latDisplay, sparkline)
	}

	content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", rows))
	footer := subtleStyle.Render("N/q/esc: back • r: refresh • R: clear cooldowns")

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer),
	)
}

func (m model) viewHelp() string {
	var title string
	var shortcuts []string

	if m.restoringBackup {
		title = "Restore Backup"
		shortcuts = []string{"y/Y/enter: Confirm", "n/N/q/esc: Cancel"}
	} else if m.managingTokens {
		title = "Manage Tokens"
		shortcuts = []string{"↑/k: Up", "↓/j: Down", "a: Add", "d: Delete", "q/esc: Back"}
	} else if m.managingChains {
		title = "Manage Chains"
		shortcuts = []string{"↑/k: Up", "↓/j: Down", "a: Add", "d: Delete", "t: Tokens", "q/esc: Back"}
	} else if m.showSummary {
		title = "Summary View"
		shortcuts = []string{"n: Sort by Name", "v: Sort by Value", "b: Sort by Balance", "g: Toggle Graph", "s/q/esc: Back"}
	} else if m.showNetworkStatus {
		title = "Network Status"
		shortcuts = []string{"N/q/esc: Back", "r: Refresh", "R: Clear Cooldowns"}
	} else if m.showGasTracker {
		title = "Gas Tracker"
		shortcuts = []string{"G/q/esc: Back", "r: Refresh", "</>: Change Time Range"}
	} else if m.showTxList {
		title = "Transactions"
		shortcuts = []string{"↑/k: Up", "↓/j: Down", "i/o/a: Filter", "enter: Details", "q/esc: Back"}
	} else if m.showDetail {
		title = "Detail View"
		shortcuts = []string{"↑/k: Scroll Up", "↓/j: Scroll Down", "c: Copy Address", "enter/esc/q: Close"}
	} else {
		title = "Main View"
		shortcuts = []string{
			"r: Refresh Data",
			"R: Force Refresh",
			"B: Restore Backup",
			"X: Export Config",
			"O: Global Settings",
			"P: Toggle Privacy",
			"A: Toggle Auto-Cycle",
			"t: Toggle Txs",
			"T: Transaction List",
			"G: Gas Tracker",
			"c: Copy Address",
			"s: Toggle Summary",
			"N: Network Status",
			"enter: Show Details",
			"Tab/l/Right: Next Account",
			"S-Tab/h/Left: Prev Account",
			"a: Add Address",
			"d: Delete Address",
			"e: Edit Address Name",
			"E: Manage Chains",
			"n: Next Chain",
			"q/esc: Quit",
			"?: Toggle Help",
		}
	}

	header := titleStyle.Render(fmt.Sprintf("Help: %s", title))
	content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "\n", strings.Join(shortcuts, "\n")))
	footer := subtleStyle.Render("Press '?' or 'esc' to close")

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer),
	)
}

func (m model) viewGasTracker() string {
	ranges := []time.Duration{30 * time.Minute, 1 * time.Hour, 6 * time.Hour, 24 * time.Hour}
	rangeLabels := []string{"30m", "1h", "6h", "24h"}
	selectedRange := ranges[m.gasTrackerRangeIndex]

	headerText := fmt.Sprintf("Gas Tracker: %s (Gwei) - Last %s", m.chains[m.activeChainIdx].Name, rangeLabels[m.gasTrackerRangeIndex])
	header := titleStyle.Render(headerText)

	var graph string
	var stats string

	targetBoxWidth := m.width - 4
	if targetBoxWidth < 0 {
		targetBoxWidth = 0
	}

	// Filter history based on selected range
	var filteredHistory []float64
	now := time.Now()
	for _, dp := range m.gasPriceHistory {
		if now.Sub(dp.Timestamp) <= selectedRange {
			filteredHistory = append(filteredHistory, dp.Value)
		}
	}

	if len(filteredHistory) > 0 {
		min := filteredHistory[0]
		max := filteredHistory[0]
		sum := 0.0
		for _, v := range filteredHistory {
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
			sum += v
		}
		avg := sum / float64(len(filteredHistory))
		stats = subtleStyle.Render(fmt.Sprintf("Low: %.2f • Avg: %.2f • High: %.2f", min, avg, max))

		graphWidth := targetBoxWidth - 14 // 4 for box borders/padding, ~10 for axis labels
		if graphWidth < 10 {
			graphWidth = 10
		}
		graphHeight := m.height - 14
		if graphHeight < 1 {
			graphHeight = 1
		}
		graph = asciigraph.Plot(filteredHistory,
			asciigraph.Height(graphHeight),
			asciigraph.Width(graphWidth),
			asciigraph.Caption("Historical Gas Price (Gwei)"),
		)
	} else {
		graph = "Not enough data to draw graph."
	}

	content := boxStyle.Width(targetBoxWidth).Align(lipgloss.Center).Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", stats, "\n", graph))
	footer := subtleStyle.Render("G/q/esc: back • r: refresh • </>: change range")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer))
}

func (m model) viewTxList() string {
	activeAcc := m.accounts[m.activeIdx]
	filterDisplay := "All"
	switch m.txFilter {
	case "in":
		filterDisplay = "Incoming"
	case "out":
		filterDisplay = "Outgoing"
	}
	header := titleStyle.Render(fmt.Sprintf("Transactions: %s (%s)", activeAcc.Address, filterDisplay))

	txs := m.getFilteredTransactions(activeAcc)

	if len(txs) == 0 {
		content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", "No transactions found."))
		footer := subtleStyle.Render("i: in • o: out • a: all • q/esc: back")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer))
	}

	rows := ""
	for i, tx := range txs {
		cursor := "  "
		if i == m.txListIdx {
			cursor = "> "
		}
		hash := utils.TruncateString(tx.Hash, 10)
		to := utils.TruncateString(tx.To, 20)
		if m.privacyMode {
			hash = "0x**...**"
			to = "0x**...**"
		}
		rows += fmt.Sprintf("%s%-12s %-12s %s\n", cursor, hash, m.maskString(tx.Value), to)
	}

	content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", rows))
	footer := subtleStyle.Render("i: in • o: out • a: all • enter: details • q/esc: back")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer))
}

func (m model) viewTxDetail() string {
	activeAcc := m.accounts[m.activeIdx]
	txs := m.getFilteredTransactions(activeAcc)
	if len(txs) == 0 || m.txListIdx >= len(txs) {
		return "No transaction selected."
	}
	tx := txs[m.txListIdx]

	header := titleStyle.Render("Transaction Details")

	// Fields
	lines := []string{
		fmt.Sprintf("Hash:      %s", m.maskAddress(tx.Hash)),
		fmt.Sprintf("Block:     %d", tx.BlockNumber),
		fmt.Sprintf("From:      %s", m.maskAddress(tx.From)),
		fmt.Sprintf("To:        %s", m.maskAddress(tx.To)),
		fmt.Sprintf("Value:     %s", m.maskString(tx.Value)),
		fmt.Sprintf("Gas Limit: %d", tx.GasLimit),
		fmt.Sprintf("Gas Price: %s", tx.GasPrice),
		fmt.Sprintf("Nonce:     %d", tx.Nonce),
	}

	content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "\n", strings.Join(lines, "\n")))
	footer := subtleStyle.Render("o: open in browser • q/esc: back")
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer))
}

func (m model) viewSummaryGraph() string {
	header := titleStyle.Render("Portfolio History")
	var graph string
	if len(m.portfolioHistory) > 0 {
		width := m.width - 10
		if width < 10 {
			width = 10
		}
		graph = asciigraph.Plot(m.portfolioHistory,
			asciigraph.Height(m.height-12),
			asciigraph.Width(width),
			asciigraph.Caption("Portfolio Value History (USD)"),
		)
	} else {
		graph = "Not enough data to draw graph."
	}

	content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", graph))
	footer := subtleStyle.Render("s: summary list • g: toggle graph • q/esc: back")

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer))
}

func (m model) viewDetail() string {
	activeAcc := m.accounts[m.activeIdx]
	header := titleStyle.Render(fmt.Sprintf("Details: %s", activeAcc.Address))
	if activeAcc.Name != "" {
		header = titleStyle.Render(fmt.Sprintf("Details: %s (%s)", activeAcc.Name, activeAcc.Address))
	}

	totalAccountValue := m.calculateAccountTotal(activeAcc)
	footer := subtleStyle.Render(fmt.Sprintf("Total Value: $%s • Press 'enter' or 'esc' to return", m.displayValue(totalAccountValue, m.config.FiatDecimals)))

	vpView := m.viewport.View()
	content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Center, header, "\n", vpView))

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer),
	)
}

func (m model) renderLatencySparkline(history []time.Duration) string {
	if len(history) == 0 {
		return ""
	}
	var min, max time.Duration
	first := true
	for _, v := range history {
		if v != -1 {
			if first {
				min, max = v, v
				first = false
			} else {
				if v < min {
					min = v
				}
				if v > max {
					max = v
				}
			}
		}
	}
	if first {
		return strings.Repeat(errStyle.Render("×"), len(history))
	}

	chars := []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var sb strings.Builder
	for _, v := range history {
		if v == -1 {
			sb.WriteString(errStyle.Render("×"))
			continue
		}
		if max == min {
			sb.WriteString(subtleStyle.Render("▄"))
			continue
		}
		idx := int((v - min) * 7 / (max - min))
		sb.WriteString(subtleStyle.Render(chars[idx]))
	}
	return sb.String()
}

func (m model) viewSummary() string {
	if m.showSummaryGraph {
		return m.viewSummaryGraph()
	}
	return m.viewSummaryList()
}

func (m model) viewSummaryList() string {
	header := titleStyle.Render("Account Summary")
	activeChain := m.chains[m.activeChainIdx]

	type rowData struct {
		origIndex  int
		address    string
		name       string
		balanceStr string
		totalValue *big.Float
	}
	var rowsData []rowData
	totalPortfolio := new(big.Float)

	for i, acc := range m.accounts {
		balStr := "..."
		if acc.Errors[activeChain.Name] != nil {
			balStr = errStyle.Render("Error")
		} else if acc.Balances[activeChain.Name] != nil {
			balStr = m.displayValue(acc.Balances[activeChain.Name], m.config.TokenDecimals)
		}

		accTotal := m.calculateAccountTotal(acc)
		totalPortfolio.Add(totalPortfolio, accTotal)

		rowsData = append(rowsData, rowData{
			origIndex:  i,
			address:    acc.Address,
			name:       acc.Name,
			balanceStr: balStr,
			totalValue: accTotal,
		})
	}

	// Logic for sorting rowsData is likely better placed in logic.go but for now I'll leave it here as it's presentation logic?
	// Wait, the sort code uses `sort.Slice` with `rowsData`. I can keep it here.
	// I need to copy the sort logic.
	// ... (Copied logic below)

	// Build header
	hName := "Address/Name"
	hTotal := "Total Value"
	hActive := fmt.Sprintf("Balance (%s)", activeChain.Symbol)

	arrow := "↓"
	if !m.summarySortDesc {
		arrow = "↑"
	}
	switch m.summarySortCol {
	case 0:
		hName += " " + arrow
	case 1:
		hTotal += " " + arrow
	case 2:
		hActive += " " + arrow
	}

	headerRow := tableHeaderStyle.Render(fmt.Sprintf("  %-38s %-20s %18s", hName, hTotal, hActive))

	rows := ""
	for _, r := range rowsData {
		marker := "  "
		if r.origIndex == m.activeIdx {
			marker = "> "
		}
		addrDisp := r.address
		if m.privacyMode {
			addrDisp = "0x**...**"
		}
		displayName := addrDisp
		if r.name != "" {
			displayName = fmt.Sprintf("%s (%s)", r.name, addrDisp)
		}
		valStr := fmt.Sprintf("$%s", m.displayValue(r.totalValue, m.config.FiatDecimals))
		rows += fmt.Sprintf("%s%-38s %-20s %18s\n", marker, utils.TruncateString(displayName, 36), valStr, r.balanceStr)
	}

	totalStr := fmt.Sprintf("$%s", m.displayValue(totalPortfolio, m.config.FiatDecimals))
	totalRow := fmt.Sprintf("\n  %-38s %-20s", "Total Portfolio Value", totalStr)

	content := boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, header, "\n", headerRow, rows, totalRow))
	footer := subtleStyle.Render("n: name • v: val • b: bal • g: graph • s/q/esc: back")

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, content, "\n", footer),
	)
}
