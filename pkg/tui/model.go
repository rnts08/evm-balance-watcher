package tui

import (
	"math/big"
	"strings"
	"time"

	"evmbal/pkg/config"
	"evmbal/pkg/models"
	"evmbal/pkg/watcher"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Version is set by Start()
var Version = "dev"

// --- Messages ---

type clearStatusMsg struct{}
type uiTickMsg time.Time
type privacyTimeoutMsg struct{}
type autoCycleMsg struct{}

// --- Model ---

type model struct {
	chains                 []config.ChainConfig
	activeChainIdx         int
	prices                 map[string]float64 // Key: CoinGecko ID
	gasPrice               *big.Int
	gasTrend               int
	accounts               []*models.Account
	activeIdx              int
	width                  int
	height                 int
	loading                bool
	lastUpdate             time.Time
	spinner                spinner.Model
	statusMessage          string
	showSummary            bool
	addressInputs          []textinput.Model
	adding                 bool
	configPath             string
	managingChains         bool
	chainListIdx           int
	addingChain            bool
	chainInputs            []textinput.Model
	managingTokens         bool
	tokenListIdx           int
	addingToken            bool
	tokenInputs            []textinput.Model
	selectedChainForTokens int
	portfolioHistory       []float64
	editingAddress         bool
	editAddressInput       textinput.Model
	rpcCooldowns           map[string]time.Time
	showNetworkStatus      bool
	rpcLatencies           map[string]time.Duration
	rpcLatencyHistory      map[string][]time.Duration
	showDetail             bool
	viewport               viewport.Model
	restoringBackup        bool
	showHelp               bool
	exportingConfig        bool
	exportInput            textinput.Model
	compactMode            bool
	showSummaryGraph       bool
	summarySortCol         int // 0: Name, 1: Value, 2: Balance
	summarySortDesc        bool
	gasPriceHistory        []models.GasPricePoint
	showGasTracker         bool
	gasTrackerRangeIndex   int // 0: 30m, 1: 1h, 2: 6h, 3: 24h
	privacyMode            bool
	lastInteraction        time.Time
	config                 config.GlobalConfig
	editingGlobalConfig    bool
	globalConfigInputs     []textinput.Model
	showTxList             bool
	txListIdx              int
	showTxDetail           bool
	txFilter               string // "all", "in", "out"
	nextAutoCycleTime      time.Time
	watcher                *watcher.Watcher
}

func initialModel(w *watcher.Watcher, addresses []config.AddressConfig, chains []config.ChainConfig, activeChainIdx int, globalCfg config.GlobalConfig, configPath string) model {
	var accounts []*models.Account
	for _, a := range addresses {
		clean := strings.TrimSpace(a.Address)
		if clean != "" {
			accounts = append(accounts, &models.Account{
				Address:       clean,
				Name:          a.Name,
				Balances:      make(map[string]*big.Float),
				TokenBalances: make(map[string]map[string]*big.Float),
				Balances24h:   make(map[string]*big.Float),
				Errors:        make(map[string]error),
			})
		}
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	ais := make([]textinput.Model, 2)
	for i := range ais {
		ais[i] = textinput.New()
		ais[i].Width = 40
	}
	ais[0].Placeholder = "0x..."
	ais[1].Placeholder = "Tag/Name (Optional)"

	cis := make([]textinput.Model, 5)
	for i := range cis {
		cis[i] = textinput.New()
		cis[i].Width = 50
	}
	cis[0].Placeholder = "Name (e.g. Optimism)"
	cis[1].Placeholder = "Symbol (e.g. ETH)"
	cis[2].Placeholder = "CoinGecko ID (e.g. optimism)"
	cis[3].Placeholder = "RPC URLs (comma separated)"
	cis[4].Placeholder = "Explorer URL (e.g. https://etherscan.io)"

	tis := make([]textinput.Model, 4)
	for i := range tis {
		tis[i] = textinput.New()
		tis[i].Width = 50
	}
	tis[0].Placeholder = "Symbol (e.g. USDC)"
	tis[1].Placeholder = "Address (0x...)"
	tis[2].Placeholder = "Decimals (e.g. 6)"
	tis[3].Placeholder = "CoinGecko ID (e.g. usd-coin)"

	editTi := textinput.New()
	editTi.Placeholder = "Tag/Name"
	editTi.Width = 40

	exportTi := textinput.New()
	exportTi.Placeholder = "/path/to/config.json"
	exportTi.Width = 50

	gcis := make([]textinput.Model, 5)
	for i := range gcis {
		gcis[i] = textinput.New()
		gcis[i].Width = 10
	}

	vp := viewport.New(0, 0)

	return model{
		accounts:             accounts,
		chains:               chains,
		activeChainIdx:       activeChainIdx,
		loading:              true,
		spinner:              s,
		addressInputs:        ais,
		configPath:           configPath,
		chainInputs:          cis,
		tokenInputs:          tis,
		prices:               make(map[string]float64),
		editAddressInput:     editTi,
		rpcCooldowns:         make(map[string]time.Time),
		rpcLatencies:         make(map[string]time.Duration),
		rpcLatencyHistory:    make(map[string][]time.Duration),
		showDetail:           false,
		viewport:             vp,
		restoringBackup:      false,
		showHelp:             false,
		exportingConfig:      false,
		exportInput:          exportTi,
		compactMode:          true,
		showSummaryGraph:     false,
		summarySortCol:       1,
		summarySortDesc:      true,
		gasPriceHistory:      make([]models.GasPricePoint, 0),
		showGasTracker:       false,
		gasTrackerRangeIndex: 0,
		privacyMode:          false,
		lastInteraction:      time.Now(),
		config:               globalCfg,
		globalConfigInputs:   gcis,
		showTxList:           false,
		txListIdx:            0,
		showTxDetail:         false,
		txFilter:             "all",
		nextAutoCycleTime:    time.Now(),
		watcher:              w,
	}
}

func (m model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Subscribe to watcher events
	cmds = append(cmds, listenForWatcher(m.watcher.Subscribe()))

	m.spinner.Tick()
	cmds = append(cmds, m.spinner.Tick)

	if !m.privacyMode && m.config.PrivacyTimeoutSeconds > 0 {
		cmds = append(cmds, tea.Tick(time.Duration(m.config.PrivacyTimeoutSeconds)*time.Second, func(t time.Time) tea.Msg {
			return privacyTimeoutMsg{}
		}))
	}

	if m.config.AutoCycleEnabled && m.config.AutoCycleIntervalSeconds > 0 {
		interval := time.Duration(m.config.AutoCycleIntervalSeconds) * time.Second
		cmds = append(cmds, tea.Tick(interval, func(t time.Time) tea.Msg {
			return autoCycleMsg{}
		}))
	}
	cmds = append(cmds, tea.Tick(time.Second, func(t time.Time) tea.Msg { return uiTickMsg(t) }))
	return tea.Batch(cmds...)
}
