package watcher

import (
	"context"
	"math/big"
	"testing"
	"time"

	"evmbal/pkg/config"
	"evmbal/pkg/models"
	"evmbal/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDataSource struct {
	mock.Mock
}

func (m *MockDataSource) FetchEthPrice(coinID string) (models.PriceData, error) {
	args := m.Called(coinID)
	return args.Get(0).(models.PriceData), args.Error(1)
}

func (m *MockDataSource) FetchChainData(chain config.ChainConfig, accounts []*models.Account) (models.ChainData, error) {
	args := m.Called(chain, accounts)
	return args.Get(0).(models.ChainData), args.Error(1)
}

func (m *MockDataSource) FetchGasPrice(rpcURLs []string) (models.GasPriceData, error) {
	args := m.Called(rpcURLs)
	return args.Get(0).(models.GasPriceData), args.Error(1)
}

func (m *MockDataSource) FetchTransactions(address string, rpcURLs []string, decimals int) ([]models.Transaction, []string, error) {
	args := m.Called(address, rpcURLs, decimals)
	return args.Get(0).([]models.Transaction), args.Get(1).([]string), args.Error(2)
}

func TestNewWatcher(t *testing.T) {
	addresses := []config.AddressConfig{{Address: "0x123", Name: "Test"}}
	chains := []config.ChainConfig{{Name: "Eth", Symbol: "ETH"}}
	globalCfg := config.GlobalConfig{}

	w := NewWatcher(addresses, chains, globalCfg, "")

	assert.NotNil(t, w)
	assert.Equal(t, 1, len(w.GetAccounts()))
	assert.Equal(t, "0x123", w.GetAccounts()[0].Address)
}

func TestSubscribeUnsubscribe(t *testing.T) {
	w := NewWatcher(nil, nil, config.GlobalConfig{}, "")
	sub := w.Subscribe()
	assert.NotNil(t, sub)

	w.mu.RLock()
	assert.Equal(t, 1, len(w.subscribers))
	w.mu.RUnlock()

	w.Unsubscribe(sub)
	w.mu.RLock()
	assert.Equal(t, 0, len(w.subscribers))
	w.mu.RUnlock()
}

func TestFetchAll(t *testing.T) {
	mockDS := new(MockDataSource)
	addresses := []config.AddressConfig{{Address: "0x123", Name: "Test"}}
	chains := []config.ChainConfig{{Name: "Eth", Symbol: "ETH", CoinGeckoID: "ethereum"}}
	globalCfg := config.GlobalConfig{TokenDecimals: 18}

	w := NewWatcher(addresses, chains, globalCfg, "")
	w.SetDataSource(mockDS)

	// Setup expectations
	mockDS.On("FetchEthPrice", "ethereum").Return(models.PriceData{CoinID: "ethereum", Price: 2000.0}, nil)
	mockDS.On("FetchChainData", mock.Anything, mock.Anything).Return(models.ChainData{
		ChainName: "Eth",
		Results: []models.AccountChainData{
			{Address: "0x123", Balance: big.NewFloat(1.5)},
		},
	}, nil)
	mockDS.On("FetchGasPrice", mock.Anything).Return(models.GasPriceData{Price: big.NewInt(20000000000)}, nil)
	mockDS.On("FetchTransactions", "0x123", mock.Anything, 18).Return([]models.Transaction{}, []string{}, nil)

	sub := w.Subscribe()

	w.fetchAll()

	mockDS.AssertExpectations(t)

	// Check state
	assert.Equal(t, 2000.0, w.GetPrices()["ethereum"])
	acc := w.GetAccounts()[0]
	assert.Equal(t, 1.5, utils.BigFloatToFloat64(acc.Balances["Eth"]))

	// Check notifications
	timeout := time.After(1 * time.Second)
	eventsCount := 0
	for i := 0; i < 4; i++ {
		select {
		case <-sub:
			eventsCount++
		case <-timeout:
			t.Errorf("Timed out waiting for events, got %d", eventsCount)
			return
		}
	}
	assert.Equal(t, 4, eventsCount)
}

func TestPollingLoop(t *testing.T) {
	mockDS := new(MockDataSource)
	w := NewWatcher(nil, nil, config.GlobalConfig{}, "")
	w.SetDataSource(mockDS)

	// Expect at least one fetchAll
	mockDS.On("FetchEthPrice", mock.Anything).Return(models.PriceData{}, nil).Maybe()
	mockDS.On("FetchChainData", mock.Anything, mock.Anything).Return(models.ChainData{}, nil).Maybe()
	mockDS.On("FetchGasPrice", mock.Anything).Return(models.GasPriceData{}, nil).Maybe()
	mockDS.On("FetchTransactions", mock.Anything, mock.Anything, mock.Anything).Return([]models.Transaction{}, []string{}, nil).Maybe()

	ctx, cancel := context.WithCancel(context.Background())
	go w.Start(ctx)

	time.Sleep(100 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
}
