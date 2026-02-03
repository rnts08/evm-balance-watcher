package tui

import (
	"math/big"
	"testing"

	"evmbal/pkg/config"
	"evmbal/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestCalculateTotalPortfolioValue(t *testing.T) {
	m := model{
		chains: []config.ChainConfig{
			{Name: "Eth", CoinGeckoID: "ethereum", Symbol: "ETH"},
		},
		prices: map[string]float64{
			"ethereum": 2000.0,
		},
		accounts: []*models.Account{
			{
				Address:  "0x123",
				Balances: map[string]*big.Float{"Eth": big.NewFloat(1.5)},
			},
		},
	}

	val := m.calculateTotalPortfolioValue()
	assert.Equal(t, 3000.0, val)
}

func TestGetFilteredTransactions(t *testing.T) {
	acc := &models.Account{
		Address: "0x123",
		Transactions: []models.Transaction{
			{From: "0x123", To: "0xabc", Value: "1.0"},
			{From: "0xdef", To: "0x123", Value: "2.0"},
		},
	}

	m := model{txFilter: "all"}
	txs := m.getFilteredTransactions(acc)
	assert.Equal(t, 2, len(txs))

	m.txFilter = "out"
	txs = m.getFilteredTransactions(acc)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, "0x123", txs[0].From)

	m.txFilter = "in"
	txs = m.getFilteredTransactions(acc)
	assert.Equal(t, 1, len(txs))
	assert.Equal(t, "0xdef", txs[0].From)
}
