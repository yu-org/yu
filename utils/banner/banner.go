package banner

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yu-org/yu/config"
)

// ANSI styles used by the chain-spec banner.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	italic = "\033[3m"

	fgCyan  = "\033[96m"
	fgWhite = "\033[97m"
	fgGreen = "\033[92m"
	fgGray  = "\033[90m"

	// badges of the network category
	bgRed     = "\033[41;97m"
	bgYellow  = "\033[43;30m"
	bgBlue    = "\033[44;97m"
	bgMagenta = "\033[45;97m"
)

const indent = "  "

// PrintChainSpec prints the identity of the chain out to os.Stdout.
// It is the very first thing shown when the chain runs.
func PrintChainSpec(spec *config.ChainSpec) {
	FprintChainSpec(os.Stdout, spec, colorEnabled(os.Stdout))
}

// FprintChainSpec writes the banner to w, colorful only if color is true.
func FprintChainSpec(w io.Writer, spec *config.ChainSpec, color bool) {
	s := *spec
	s.FillDefaults()

	c := painter(color)

	bigName := RenderBig(s.ChainName)
	width := 0
	for _, line := range bigName {
		width = max(width, len([]rune(line)))
	}
	// keep room for the info lines below
	width = max(width, 34)
	rule := strings.Repeat("─", width)

	fmt.Fprintln(w)
	for _, line := range bigName {
		fmt.Fprintln(w, indent+c(bold+fgCyan, line))
	}
	fmt.Fprintln(w, indent+c(fgGray, rule))
	fmt.Fprintln(w, indent+c(dim, "Author  ")+c(bold+fgWhite, s.Author))
	fmt.Fprintln(w, indent+c(dim, "Version ")+c(fgGreen, s.Version))
	fmt.Fprintln(w, indent+c(dim, "Network ")+networkBadge(s.Network, c))
	fmt.Fprintln(w, indent+c(fgGray, rule))
	fmt.Fprintln(w)
}

func networkBadge(network config.Network, c func(string, string) string) string {
	style := bgMagenta
	switch strings.ToLower(network) {
	case config.Mainnet:
		style = bgRed
	case config.Testnet:
		style = bgYellow
	case config.Devnet:
		style = bgBlue
	case config.Localnet:
		style = bgMagenta
	}
	badge := c(bold+style, " "+strings.ToUpper(network)+" ")
	if network == config.Mainnet {
		return badge
	}
	return badge + " " + c(dim+italic, "(not the main network)")
}

// painter returns the func styling a text, it does nothing when color is off.
func painter(color bool) func(style, text string) string {
	if !color {
		return func(_, text string) string { return text }
	}
	return func(style, text string) string { return style + text + reset }
}

// colorEnabled reports whether ANSI styles work on w.
func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	// only a terminal understands the ANSI escape codes
	return stat.Mode()&os.ModeCharDevice != 0
}
