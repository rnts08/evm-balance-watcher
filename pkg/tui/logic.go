package tui

import (
	"fmt"
	"math/big"
	"strings"

	"evmbal/pkg/models"
	"evmbal/pkg/watcher"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m model) calculateTotalPortfolioValue() float64 {
	total := new(big.Float)
	for _, acc := range m.accounts {
		for _, chain := range m.chains {
			if bal, ok := acc.Balances[chain.Name]; ok {
				if price, ok := m.prices[chain.CoinGeckoID]; ok {
					val := new(big.Float).Mul(bal, big.NewFloat(price))
					total.Add(total, val)
				}
			}
			if tokens, ok := acc.TokenBalances[chain.Name]; ok {
				for _, t := range chain.Tokens {
					if bal, ok := tokens[t.Symbol]; ok {
						if price, ok := m.prices[t.CoinGeckoID]; ok {
							val := new(big.Float).Mul(bal, big.NewFloat(price))
							total.Add(total, val)
						}
					}
				}
			}
		}
	}
	f, _ := total.Float64()
	return f
}

func (m *model) updateDetailViewport() {
	activeAcc := m.accounts[m.activeIdx]
	var sections []string

	for _, chain := range m.chains {
		// Only show chains with balances or tokens
		hasContent := false
		chainTotal := new(big.Float)
		var itemRows []string

		// Native Balance
		if bal, ok := activeAcc.Balances[chain.Name]; ok {
			val := new(big.Float)
			price := m.prices[chain.CoinGeckoID]
			if price > 0 {
				val = new(big.Float).Mul(bal, big.NewFloat(price))
			}
			chainTotal.Add(chainTotal, val)

			valStr := ""
			if price > 0 {
				valStr = fmt.Sprintf("($%s)", m.displayValue(val, m.config.FiatDecimals))
			}
			itemRows = append(itemRows, fmt.Sprintf("  %-8s %12s %s", chain.Symbol, m.displayValue(bal, m.config.TokenDecimals), valStr))
			hasContent = true
		}

		// Token Balances
		if tokens, ok := activeAcc.TokenBalances[chain.Name]; ok {
			for _, t := range chain.Tokens {
				if bal, ok := tokens[t.Symbol]; ok && bal.Sign() > 0 {
					val := new(big.Float)
					price := m.prices[t.CoinGeckoID]
					if price > 0 {
						val = new(big.Float).Mul(bal, big.NewFloat(price))
					}
					chainTotal.Add(chainTotal, val)

					valStr := ""
					if price > 0 {
						valStr = fmt.Sprintf("($%s)", m.displayValue(val, m.config.FiatDecimals))
					}
					itemRows = append(itemRows, fmt.Sprintf("  %-8s %12s %s", t.Symbol, m.displayValue(bal, m.config.TokenDecimals), valStr))
					hasContent = true
				}
			}
		}

		if hasContent {
			chainHeader := fmt.Sprintf("%s (Total: $%s)", chain.Name, m.displayValue(chainTotal, m.config.FiatDecimals))
			section := lipgloss.JoinVertical(lipgloss.Left,
				subtleStyle.Render(chainHeader),
				strings.Join(itemRows, "\n"),
			)
			sections = append(sections, section)
		}
	}

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	if len(sections) == 0 {
		content = "No balances found."
	}
	m.viewport.SetContent(content)
}

func (m model) calculateAccountTotal(acc *models.Account) *big.Float {
	total := new(big.Float)
	for _, chain := range m.chains {
		if bal, ok := acc.Balances[chain.Name]; ok {
			if price, ok := m.prices[chain.CoinGeckoID]; ok {
				val := new(big.Float).Mul(bal, big.NewFloat(price))
				total.Add(total, val)
			}
		}
		if tokens, ok := acc.TokenBalances[chain.Name]; ok {
			for _, t := range chain.Tokens {
				if bal, ok := tokens[t.Symbol]; ok {
					if price, ok := m.prices[t.CoinGeckoID]; ok {
						val := new(big.Float).Mul(bal, big.NewFloat(price))
						total.Add(total, val)
					}
				}
			}
		}
	}
	return total
}

func (m model) getFilteredTransactions(acc *models.Account) []models.Transaction {
	if m.txFilter == "all" || m.txFilter == "" {
		return acc.Transactions
	}
	var filtered []models.Transaction
	for _, tx := range acc.Transactions {
		isFrom := strings.EqualFold(tx.From, acc.Address)
		if m.txFilter == "in" && !isFrom {
			filtered = append(filtered, tx)
		} else if m.txFilter == "out" && isFrom {
			filtered = append(filtered, tx)
		}
	}
	return filtered
}

func listenForWatcher(sub watcher.Subscriber) tea.Cmd {
	return func() tea.Msg {
		return <-sub
	}
}
