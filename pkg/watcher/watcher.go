package watcher

import (
	"context"
	"math/big"
	"sync"
	"time"

	"evmbal/pkg/config"
	"evmbal/pkg/models"
	"evmbal/pkg/rpc"
)

// DataSource defines the interface for fetching data.
type DataSource interface {
	FetchEthPrice(coinID string) (models.PriceData, error)
	FetchChainData(chain config.ChainConfig, accounts []*models.Account) (models.ChainData, error)
	FetchGasPrice(rpcURLs []string) (models.GasPriceData, error)
	FetchTransactions(address string, rpcURLs []string, decimals int) ([]models.Transaction, []string, error)
}

// RealDataSource implements DataSource using the rpc package.
type RealDataSource struct{}

func (d *RealDataSource) FetchEthPrice(coinID string) (models.PriceData, error) {
	return rpc.FetchEthPrice(coinID)
}

func (d *RealDataSource) FetchChainData(chain config.ChainConfig, accounts []*models.Account) (models.ChainData, error) {
	return rpc.FetchChainData(chain, accounts)
}

func (d *RealDataSource) FetchGasPrice(rpcURLs []string) (models.GasPriceData, error) {
	return rpc.FetchGasPrice(rpcURLs)
}

func (d *RealDataSource) FetchTransactions(address string, rpcURLs []string, decimals int) ([]models.Transaction, []string, error) {
	return rpc.FetchTransactions(address, rpcURLs, decimals)
}

// Watcher manages background monitoring and state.
type Watcher struct {
	config     config.GlobalConfig
	addresses  []config.AddressConfig
	chains     []config.ChainConfig
	configPath string

	prices    map[string]float64
	gasPrices map[string]*big.Int
	accounts  []*models.Account

	subscribers []Subscriber
	mu          sync.RWMutex
	stopChan    chan struct{}
	dataSource  DataSource
}

// NewWatcher creates a new Watcher instance.
func NewWatcher(addresses []config.AddressConfig, chains []config.ChainConfig, globalCfg config.GlobalConfig, configPath string) *Watcher {
	var accounts []*models.Account
	for _, a := range addresses {
		accounts = append(accounts, &models.Account{
			Address:       a.Address,
			Name:          a.Name,
			Balances:      make(map[string]*big.Float),
			TokenBalances: make(map[string]map[string]*big.Float),
			Balances24h:   make(map[string]*big.Float),
			Errors:        make(map[string]error),
		})
	}

	return &Watcher{
		config:     globalCfg,
		addresses:  addresses,
		chains:     chains,
		configPath: configPath,
		prices:     make(map[string]float64),
		gasPrices:  make(map[string]*big.Int),
		accounts:   accounts,
		stopChan:   make(chan struct{}),
		dataSource: &RealDataSource{},
	}
}

// SetDataSource allows overriding the data source (useful for testing).
func (w *Watcher) SetDataSource(ds DataSource) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.dataSource = ds
}

// Subscribe adds a new subscriber and returns a channel to receive events.
func (w *Watcher) Subscribe() Subscriber {
	w.mu.Lock()
	defer w.mu.Unlock()
	ch := make(Subscriber, 100)
	w.subscribers = append(w.subscribers, ch)
	return ch
}

// Unsubscribe removes a subscriber.
func (w *Watcher) Unsubscribe(ch Subscriber) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, sub := range w.subscribers {
		if sub == ch {
			w.subscribers = append(w.subscribers[:i], w.subscribers[i+1:]...)
			close(ch)
			break
		}
	}
}

func (w *Watcher) notify(event Event) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	for _, sub := range w.subscribers {
		select {
		case sub <- event:
		default:
			// Subscriber is slow, skip or handle?
		}
	}
}

// Start begins the monitoring loops.
func (w *Watcher) Start(ctx context.Context) {
	go w.pollingLoop(ctx)
}

// Stop stops the monitoring loops.
func (w *Watcher) Stop() {
	close(w.stopChan)
}

func (w *Watcher) pollingLoop(ctx context.Context) {
	// Initial fetch
	w.fetchAll()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.fetchAll()
		case <-w.stopChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (w *Watcher) fetchAll() {
	var wg sync.WaitGroup

	// Fetch Prices
	uniqueCoinIDs := make(map[string]bool)
	for _, chain := range w.chains {
		if chain.CoinGeckoID != "" {
			uniqueCoinIDs[chain.CoinGeckoID] = true
		}
		for _, t := range chain.Tokens {
			if t.CoinGeckoID != "" {
				uniqueCoinIDs[t.CoinGeckoID] = true
			}
		}
	}

	for id := range uniqueCoinIDs {
		wg.Add(1)
		go func(coinID string) {
			defer wg.Done()
			data, err := w.dataSource.FetchEthPrice(coinID)
			if err == nil {
				w.mu.Lock()
				w.prices[coinID] = data.Price
				w.mu.Unlock()
				w.notify(Event{Type: EventPriceUpdated, Data: data})
			}
		}(id)
	}

	// Fetch Chain Data (Balances)
	for _, chain := range w.chains {
		wg.Add(1)
		go func(c config.ChainConfig) {
			defer wg.Done()
			data, err := w.dataSource.FetchChainData(c, w.accounts)
			if err == nil {
				w.updateAccountsWithChainData(data)
				w.notify(Event{Type: EventChainDataUpdated, Data: data})
			}
		}(chain)

		wg.Add(1)
		go func(c config.ChainConfig) {
			defer wg.Done()
			data, err := w.dataSource.FetchGasPrice(c.RPCURLs)
			if err == nil {
				w.mu.Lock()
				w.gasPrices[c.Name] = data.Price
				w.mu.Unlock()
				w.notify(Event{Type: EventGasPriceUpdated, Data: data})
			}
		}(chain)

		for _, acc := range w.accounts {
			wg.Add(1)
			go func(c config.ChainConfig, address string) {
				defer wg.Done()
				txs, _, err := w.dataSource.FetchTransactions(address, c.RPCURLs, w.config.TokenDecimals)
				if err == nil {
					w.mu.Lock()
					for _, a := range w.accounts {
						if a.Address == address {
							a.Transactions = txs
							break
						}
					}
					w.mu.Unlock()
					w.notify(Event{Type: EventTransactionsUpdated, Data: map[string]interface{}{
						"address": address,
						"txs":     txs,
					}})
				}
			}(chain, acc.Address)
		}
	}

	wg.Wait()
}

func (w *Watcher) updateAccountsWithChainData(data models.ChainData) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, res := range data.Results {
		for _, acc := range w.accounts {
			if acc.Address == res.Address {
				if acc.Balances == nil {
					acc.Balances = make(map[string]*big.Float)
				}
				if acc.TokenBalances == nil {
					acc.TokenBalances = make(map[string]map[string]*big.Float)
				}
				acc.Balances[data.ChainName] = res.Balance
				acc.Balances24h[data.ChainName] = res.Balance24h
				if acc.TokenBalances[data.ChainName] == nil {
					acc.TokenBalances[data.ChainName] = make(map[string]*big.Float)
				}
				for sym, bal := range res.TokenBalances {
					acc.TokenBalances[data.ChainName][sym] = bal
				}
				break
			}
		}
	}
}

// GetAccounts returns a copy of the current accounts state.
func (w *Watcher) GetAccounts() []*models.Account {
	w.mu.RLock()
	defer w.mu.RUnlock()
	// Deep copy would be better but for now just return the slice
	return w.accounts
}

// GetPrices returns the current prices.
func (w *Watcher) GetPrices() map[string]float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	cp := make(map[string]float64)
	for k, v := range w.prices {
		cp[k] = v
	}
	return cp
}
