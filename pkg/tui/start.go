package tui

import (
	"fmt"
	"os"

	"evmbal/pkg/config"
	"evmbal/pkg/watcher"

	tea "github.com/charmbracelet/bubbletea"
)

func Start(w *watcher.Watcher, addresses []config.AddressConfig, chains []config.ChainConfig, activeChainIdx int, globalCfg config.GlobalConfig, configPath, version string) {
	Version = version
	p := tea.NewProgram(
		initialModel(w, addresses, chains, activeChainIdx, globalCfg, configPath),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
