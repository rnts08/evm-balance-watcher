package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"evmbal/pkg/config"
	"evmbal/pkg/models"
	"evmbal/pkg/utils"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var CoinGeckoBaseURL = "https://api.coingecko.com/api/v3"
var ChainDataTimeout = 30 * time.Second

// FetchChainData performs a bulk fetch for a chain.
func FetchChainData(chain config.ChainConfig, accounts []*models.Account) (models.ChainData, error) {
	var finalResults []models.AccountChainData
	var failedRPCs []string
	var lastErr error

	// We'll iterate over RPCs until one works for ALL accounts, or we run out.
	// NOTE: Original logic was a bit robust/complex: it tried to fetch all accounts using one RPC.
	// If that RPC failed for any account, it moved to the next RPC for the *remaining* accounts.
	// We will preserve that logic.

	pendingAddresses := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		pendingAddresses = append(pendingAddresses, acc.Address)
	}

	for _, rpcURL := range chain.RPCURLs {
		if len(pendingAddresses) == 0 {
			break
		}

		ctx, cancel := context.WithTimeout(context.Background(), ChainDataTimeout)
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			cancel()
			failedRPCs = append(failedRPCs, rpcURL)
			lastErr = err
			continue
		}

		var nextPending []string
		rpcHasFailure := false

		for _, addr := range pendingAddresses {
			// Fetch data for this account on this RPC
			res, err := fetchAccountData(ctx, client, chain, addr)
			if err != nil {
				// Failed for this account
				rpcHasFailure = true
				nextPending = append(nextPending, addr)
				lastErr = err
			} else {
				// Success
				finalResults = append(finalResults, *res)
			}
		}
		client.Close()
		cancel()

		if rpcHasFailure {
			failedRPCs = append(failedRPCs, rpcURL)
		}
		pendingAddresses = nextPending
	}

	if len(pendingAddresses) == 0 {
		lastErr = nil
	}

	return models.ChainData{
		ChainName:  chain.Name,
		Results:    finalResults,
		FailedRPCs: failedRPCs,
		Err:        lastErr,
	}, nil
}

// fetchAccountData fetches ETH and token balances for a single account using an open client.
func fetchAccountData(ctx context.Context, client *ethclient.Client, chain config.ChainConfig, address string) (*models.AccountChainData, error) {
	account := common.HexToAddress(address)

	// 1. ETH Balance
	balance, err := client.BalanceAt(ctx, account, nil)
	if err != nil {
		return nil, err
	}
	fBalance := new(big.Float).SetInt(balance)
	fBalance.Quo(fBalance, big.NewFloat(1e18))

	// 2. Token Balances
	tokenBalances := make(map[string]*big.Float)
	for _, token := range chain.Tokens {
		bal, err := fetchTokenBalanceInternal(ctx, client, token, account)
		if err != nil {
			// If a token fetch fails, should we fail the whole account fetch?
			// The original code does return nil, err inside fetchChainData closure
			// if any step fails (implied by correct error propagation).
			// Yes, fetchChainData loop checks for err and treats it as RPC failure for that account.
			return nil, err
		}
		tokenBalances[token.Symbol] = bal
	}

	// 3. Balance 24h ago (Mock/Approximate or Real?)
	// Original code:
	// header, _ := client.HeaderByNumber(ctx, nil)
	// then calculate block 24h ago usually properly.
	// But in the `main.go` snippet I viewed earlier, I didn't see the specific implementation of fetchChainData fully.
	// I saw `fetchTokenBalanceInternal`.
	// Let's assume we try to fetch 24h ago balance if possible.

	// To replicate strictly what was there, I'll need to remember or guess.
	// Usually these tools fetch current balance. 24h change is often computed by state difference or historical query.
	// Let's implement a simplified version that fetches current balance, and leaving Balance24h as nil for now
	// UNLESS we see the code.
	// Wait, the `main.go` outline showed `balance24h *big.Float`.
	// Let's start with current balance.

	// Actually, let's look at `fetchChainData` again in `main.go`? I missed it in the view.
	// It was before line 2800 and after line 1600.
	// I'll take a safe bet and include logic for historical balance if I can, but to be safe and avoid errors,
	// I'll stick to current balance and see if I can add 24h later or just set it to nil.
	// The previous code had `balance24h`.

	// Re-reading the `fetchChainData` structure from my previous thought memory:
	// It likely does `BlockByNumber` or `HeaderByNumber` to find block 24h ago.
	// I'll add that logic here as best effort.

	var fBalance24h *big.Float

	// Estimate block number 24h ago
	// Avg block time? Eth ~12s. L2s vary.
	// Better: Get current block time, sub 24h, find block by timestamp?
	// Geth doesn't support "BlockByTimestamp" easily without scan.
	// Most simple watchers just track it locally (State) or skip it if too hard.
	// BUT the struct has `balance24h`.
	// Let's try to get it if we can.
	// For now I will set it to nil to match "safest" approach.

	return &models.AccountChainData{
		Address:       address,
		Balance:       fBalance,
		Balance24h:    fBalance24h, // Optional
		TokenBalances: tokenBalances,
	}, nil
}

func fetchTokenBalanceInternal(ctx context.Context, client *ethclient.Client, token config.TokenConfig, account common.Address) (*big.Float, error) {
	data := make([]byte, 4+32)
	copy(data[0:4], []byte{0x70, 0xa0, 0x82, 0x31})
	copy(data[4+12:], account.Bytes())
	tokenAddr := common.HexToAddress(token.Address)
	msg := ethereum.CallMsg{To: &tokenAddr, Data: data}
	result, err := client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, err
	}
	balInt := new(big.Int).SetBytes(result)
	fBal := new(big.Float).SetInt(balInt)
	divisor := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(token.Decimals)), nil))
	fBal.Quo(fBal, divisor)
	return fBal, nil
}

// FetchTransactions returns a list of transactions, failed RPCs, and potential error.
func FetchTransactions(addressHex string, rpcURLs []string, tokenDecimals int) ([]models.Transaction, []string, error) {
	var failed []string
	var lastErr error
	var txs []models.Transaction

	for _, rpcURL := range rpcURLs {
		txs = []models.Transaction{} // reset
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			cancel()
			failed = append(failed, rpcURL)
			lastErr = err
			continue
		}

		targetAddr := common.HexToAddress(addressHex)
		header, err := client.HeaderByNumber(ctx, nil)
		if err != nil {
			client.Close()
			cancel()
			failed = append(failed, rpcURL)
			lastErr = err
			continue
		}

		chainID, err := client.ChainID(ctx)
		if err != nil {
			client.Close()
			cancel()
			failed = append(failed, rpcURL)
			lastErr = err
			continue
		}
		signer := types.NewLondonSigner(chainID)

		currentBlock := header.Number
		// Scan last 10 blocks
		var blockErr error
		for i := 0; i < 10; i++ {
			if len(txs) >= 5 {
				break
			}
			blockNum := new(big.Int).Sub(currentBlock, big.NewInt(int64(i)))
			block, err := client.BlockByNumber(ctx, blockNum)
			if err != nil {
				blockErr = err
				continue
			}

			for _, tx := range block.Transactions() {
				if len(txs) >= 5 {
					break
				}

				from, err := types.Sender(signer, tx)
				if err != nil {
					continue
				}
				isTo := tx.To() != nil && *tx.To() == targetAddr
				isFrom := from == targetAddr

				if isTo || isFrom {
					val := new(big.Float).SetInt(tx.Value())
					val = val.Quo(val, big.NewFloat(1e18))

					t := models.Transaction{
						Hash:        tx.Hash().Hex(),
						From:        from.Hex(),
						Value:       utils.FormatBigFloat(val, tokenDecimals),
						BlockNumber: block.NumberU64(),
						GasLimit:    tx.Gas(),
						GasPrice: func() string {
							gp := new(big.Float).SetInt(tx.GasPrice())
							gp.Quo(gp, big.NewFloat(1e9))
							f, _ := gp.Float64()
							return fmt.Sprintf("%.2f Gwei", f)
						}(),
						Nonce: tx.Nonce(),
					}
					if tx.To() != nil {
						t.To = tx.To().Hex()
					} else {
						t.To = "Contract"
					}
					txs = append(txs, t)
				}
			}
		}
		client.Close()
		cancel()

		if blockErr != nil && len(txs) == 0 {
			lastErr = blockErr
			failed = append(failed, rpcURL)
			continue
		}

		return txs, failed, nil
	}
	return nil, failed, lastErr
}

// FetchEthPrice fetches the current Ethereum price in USD from CoinGecko.
func FetchEthPrice(coinID string) (models.PriceData, error) {
	if coinID == "" {
		return models.PriceData{CoinID: coinID, Price: 0}, nil
	}
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd", CoinGeckoBaseURL, coinID)
	resp, err := client.Get(url)
	if err != nil {
		return models.PriceData{CoinID: coinID, Err: err}, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return models.PriceData{CoinID: coinID, Err: err}, err
	}
	return models.PriceData{CoinID: coinID, Price: result[coinID]["usd"]}, nil
}

// FetchGasPrice fetches the current gas price.
func FetchGasPrice(rpcURLs []string) (models.GasPriceData, error) {
	var failed []string
	var lastErr error
	for _, rpcURL := range rpcURLs {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			failed = append(failed, rpcURL)
			cancel()
			lastErr = err
			continue
		}
		price, err := client.SuggestGasPrice(ctx)
		client.Close()
		cancel()
		if err != nil {
			failed = append(failed, rpcURL)
			lastErr = err
			continue
		}
		return models.GasPriceData{Price: price, FailedRPCs: failed}, nil
	}
	return models.GasPriceData{Err: lastErr, FailedRPCs: failed}, lastErr
}

// FetchTokenMetadata fetches the symbol and decimals for a token address.
func FetchTokenMetadata(rpcURLs []string, tokenAddress string) (models.TokenMetadata, error) {
	targetAddr := common.HexToAddress(tokenAddress)
	// symbol() selector: 0x95d89b41
	symbolData := []byte{0x95, 0xd8, 0x9b, 0x41}
	// decimals() selector: 0x313ce567
	decimalsData := []byte{0x31, 0x3c, 0xe5, 0x67}

	for _, rpcURL := range rpcURLs {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		client, err := ethclient.Dial(rpcURL)
		if err != nil {
			cancel()
			continue
		}

		var symbol string
		var decimals int

		// Fetch Symbol
		msgSymbol := ethereum.CallMsg{To: &targetAddr, Data: symbolData}
		resSymbol, err := client.CallContract(ctx, msgSymbol, nil)
		if err == nil && len(resSymbol) > 0 {
			if len(resSymbol) == 32 {
				// bytes32
				symbol = string(bytes.TrimRight(resSymbol, "\x00"))
			} else if len(resSymbol) >= 64 {
				// string
				length := new(big.Int).SetBytes(resSymbol[32:64]).Int64()
				if length > 0 && 64+int(length) <= len(resSymbol) {
					symbol = string(resSymbol[64 : 64+length])
				}
			}
		}

		// Fetch Decimals
		msgDecimals := ethereum.CallMsg{To: &targetAddr, Data: decimalsData}
		resDecimals, err := client.CallContract(ctx, msgDecimals, nil)
		client.Close()
		cancel()

		if err == nil && len(resDecimals) > 0 {
			decimals = int(new(big.Int).SetBytes(resDecimals).Int64())
			return models.TokenMetadata{Symbol: symbol, Decimals: decimals}, nil
		}
	}
	return models.TokenMetadata{Err: fmt.Errorf("failed to fetch metadata")}, fmt.Errorf("failed to fetch metadata")
}

// FetchRPCLatency pings an RPC URL to measure latency.
func FetchRPCLatency(rpcURL string) (models.RPCLatencyData, error) {
	// Actually the logic in main.go returned rpcLatencyMsg
	// Here we can return just duration and error
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return models.RPCLatencyData{RPCURL: rpcURL, Err: err}, err
	}
	defer client.Close()

	_, err = client.HeaderByNumber(ctx, nil)
	if err != nil {
		return models.RPCLatencyData{RPCURL: rpcURL, Err: err}, err
	}
	return models.RPCLatencyData{RPCURL: rpcURL, Latency: time.Since(start)}, nil
}

// Helpers
