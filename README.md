# EVM Balance Watcher

![Release Status](https://github.com/rnts08/evmbal/actions/workflows/release.yml/badge.svg)
![Test Status](https://github.com/rnts08/evmbal/actions/workflows/test.yml/badge.svg)
![Format Status](https://github.com/rnts08/evmbal/actions/workflows/format.yml/badge.svg)

A comprehensive, terminal-based UI for monitoring balances and transactions on Ethereum and other EVM-compatible chains.
A terminal user interface for tracking Ethereum and EVM chain balances.

![Screenshot](https://user-images.githubusercontent.com/12345/screenshot.png) <!-- Placeholder for screenshot -->

## Development

## Features

### Prerequisites

- Go 1.21+
- Make

- **Multi-Address & Multi-Chain:*- Monitor multiple wallet addresses across various configured EVM chains.
- **Real-time Data:*- Fetches native currency and ERC-20 token balances, USD values (via CoinGecko), and current gas prices.
- **Transaction History:*- View recent incoming and outgoing transactions for the selected address, with filtering capabilities.
- **Interactive TUI:**
  - Add, remove, and edit addresses and chains directly from the UI.
  - Add and remove ERC-20 tokens for each chain, with automatic metadata fetching (symbol, decimals).
  - Detailed views for individual accounts and transactions.
- **Network & Gas Monitoring:**
  - A dedicated "Network Status" view to check RPC latency and health.
  - A "Gas Tracker" view with a historical graph and multiple time ranges (30m, 1h, 6h, 24h).
- **Privacy & Automation:**
  - **Privacy Mode:*- Obfuscates all sensitive values and addresses, with an automatic inactivity timeout.
  - **Auto-Cycle:*- Automatically cycle through monitored addresses at a configurable interval, with a visual countdown and pause-on-interaction.
- **Robust & Configurable:**
  - Highly configurable via a `.evmbal.json` file.
  - Intelligent RPC handling with cooldowns and automatic prioritization based on latency.
  - Configuration testing, validation, and backup/restore functionality.

## Installation & Usage

### Building with Make

Ensure you have Go and `make` installed. A `Makefile` is provided to simplify common tasks.

### Building

```bash
git clone https://github.com/rnts08/evmbal.git
cd evmbal
make build
```

### Configuration

1. Create a `.evmbal.json` file. You can place it in your home directory as `~/.evmbal.json` or provide a path at runtime using the `-config` flag.
2. Use the example below as a starting point.

### Configuration example

The application is configured using a JSON file. Here is an example structure:

```json
{
  "addresses": [
    {
      "address": "0x968cC7D93c388614f620Ef812C5fdfe64029B92d",
      "name": "My Main Wallet"
    },
    {
      "address": "0x3BcA9C05410E4266C5aC1D2e544fAD7dB1fA5405"
    }
  ],
  "chains": [
    {
      "name": "Ethereum",
      "rpc_urls": [
        "https://cloudflare-eth.com",
        "https://rpc.ankr.com/eth"
      ],
      "symbol": "ETH",
      "coingecko_id": "ethereum",
      "chain_id": 1,
      "explorer_url": "https://etherscan.io",
      "tokens": [
        {
          "symbol": "USDC",
          "address": "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48",
          "decimals": 6,
          "coingecko_id": "usd-coin"
        }
      ]
    }
  ],
  "selected_chain": "Ethereum",
  "privacy_timeout_seconds": 60,
  "fiat_decimals": 2,
  "token_decimals": 2,
  "auto_cycle_enabled": false,
  "auto_cycle_interval_seconds": 15
}
```

- **`addresses`**: A list of wallet addresses to monitor. The `name` field is an optional tag.
- **`chains`**: A list of EVM chains.
  - `name`: The display name for the chain.
  - `rpc_urls`: A list of RPC endpoints. The app will prioritize them based on latency and automatically failover.
  - `symbol`: The native currency symbol (e.g., "ETH").
  - `coingecko_id`: The ID from CoinGecko's API for fetching price data.
  - `chain_id` (optional): The chain's ID, used for validation. Can be auto-populated with `-test`.
    - `explorer_url` (optional): The base URL for a block explorer, used for opening transactions in a browser.
    - `tokens`: A list of ERC-20 tokens to monitor on this chain.
- **`selected_chain`**: The name of the chain to display on startup.
- **`privacy_timeout_seconds`**: Automatically re-enable Privacy Mode after this many seconds of inactivity. Set to `0` to disable.
- **`fiat_decimals`**: Number of decimal places to show for fiat values (e.g., USD).
- **`token_decimals`**: Number of decimal places to show for token and native currency balances.
- **`auto_cycle_enabled`**: Set to `true` to automatically cycle through addresses.
- **`auto_cycle_interval_seconds`**: The delay between each address switch when auto-cycle is enabled.

### Running the Application

### Running

```bash
./evmbal
```

Or with a custom config path:

```bash
./evmbal -config /path/to/your/config.json
```

## Keybindings

### Global / Main View

| Key(s) | Action |
| :--- | :--- |
| `q`, `esc` | Quit the application. |
| `?` | Toggle the help view. |
| `r` | Refresh all data. |
| `R` | Force refresh, clearing RPC cooldowns. |
| `Tab`, `l`, `→` | Cycle to the next address. |
| `Shift+Tab`, `h`, `←` | Cycle to the previous address. |
| `n` | Cycle to the next configured chain. |
| `s` | Toggle the portfolio summary view. |
| `t` | Toggle compact mode (show/hide transactions). |
| `T` | Open the transaction list view. |
| `G` | Open the gas tracker view. |
| `N` | Open the network status view. |
| `enter` | Open the detailed view for the current address. |
| `P` | Toggle Privacy Mode. |
| `A` | Toggle auto-cycle mode. |
| `a` | Add a new address. |
| `d` | Delete the current address. |
| `e` | Edit the name/tag of the current address. |
| `E` | Open the chain management view. |
| `c` | Copy the current address to the clipboard. |
| `O` | Open the global settings editor. |
| `B` | Restore configuration from the latest backup. |
| `X` | Export the current configuration to a new file. |

### Summary View

| Key(s) | Action |
| :--- | :--- |
| `s`, `q`, `esc` | Return to the main view. |
| `g` | Toggle the portfolio history graph. |
| `n` | Sort by name. |
| `v` | Sort by total value. |
| `b` | Sort by active chain balance. |

### Transaction List View

| Key(s) | Action |
| :--- | :--- |
| `q`, `esc` | Return to the main view. |
| `↑` / `k` | Move selection up. |
| `↓` / `j` | Move selection down. |
| `i` | Filter for **i**ncoming transactions. |
| `o` | Filter for **o**utgoing transactions. |
| `a` | Filter for **a**ll transactions. |
| `enter` | View details for the selected transaction. |

### Transaction Detail View

| Key(s) | Action |
| :--- | :--- |
| `q`, `esc` | Return to the transaction list. |
| `o` | **O**pen the transaction in the chain's block explorer. |

### Gas Tracker View

| Key(s) | Action |
| :--- | :--- |
| `G`, `q`, `esc` | Return to the main view. |
| `r` | Refresh gas price. |
| `<` / `>` | Change the time range. |

### Network Status View

| Key(s) | Action |
| :--- | :--- |
| `N`, `q`, `esc` | Return to the main view. |
| `r` | Refresh latency checks. |
| `R` | Clear all RPC cooldowns. |

### Detail View (per-account)

| Key(s) | Action |
| :--- | :--- |
| `enter`, `q`, `esc` | Return to the main view. |
| `c` | Copy the account's address. |
| `↑` / `↓` | Scroll the view. |

### Management & Input Screens

| Key(s) | Action |
| :--- | :--- |
| `q`, `esc` | Cancel and return to the previous view. |
| `enter` | Move to the next field or save the form. |
| `↑` / `↓` | Move between items in a list (e.g., Manage Chains). |

## License

This project is open source and available under the GNU License.

## Tips and appreciations

***ETH/ERC20:**- 0x968cC7D93c388614f620Ef812C5fdfe64029B92d

***SOL:**- HB2o6q6vsW5796U5y7NxNqA7vYZW1vuQjpAHDo7FAMG8

***BTC:**- bc1qkmzc6d49fl0edyeynezwlrfqv486nmk6p5pmta

### Automated Release

See RELEASE.md for more information, but here is the short version:

```bash
make bump part={patch|minor|major}
```

Pushing the tag triggers a GitHub Action workflow that:

  1. Runs unit tests and configuration tests.
  2. Verifies the tag matches the `VERSION` file.
  3. Cross-compiles binaries for Linux, Windows, and macOS.
  4. Creates a GitHub Release with the binaries and an automatically generated changelog.
