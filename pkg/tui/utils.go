package tui

import (
	"os/exec"
	"runtime"

	"evmbal/pkg/utils"

	"math/big"
)

func (m model) displayValue(f *big.Float, decimals int) string {
	if m.privacyMode {
		return "****"
	}
	return utils.FormatBigFloat(f, decimals)
}

func (m model) maskString(s string) string {
	if m.privacyMode {
		return "****"
	}
	return s
}

func (m model) maskAddress(addr string) string {
	if m.privacyMode {
		return "0x**...**"
	}
	return addr
}

// openBrowser opens the specified URL in the default browser.
func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
