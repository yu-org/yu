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
	reset = "\033[0m"
	bold  = "\033[1m"
	dim   = "\033[2m"

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

const (
	indent = "  "
	// the column where the value of every info line starts, wide enough to
	// leave a gap between the label and the color block of the network badge
	labelWidth = 10
)

// label pads the name of an info line up to width.
func label(c func(style, text string) string, name string, width int) string {
	return c(dim, name+strings.Repeat(" ", width-len(name)))
}

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
	fmt.Fprintln(w, indent+label(c, "Author", labelWidth)+c(bold+fgWhite, s.Author))
	// the network badge carries a padding space of its own inside the color block,
	// so its label is one shorter to keep the three values in the same column.
	fmt.Fprintln(w, indent+label(c, "Network", labelWidth-1)+networkBadge(s.Network, c))
	fmt.Fprintln(w, indent+label(c, "Version", labelWidth)+c(fgGreen, s.Version))
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
	}
	return c(bold+style, " "+strings.ToUpper(network)+" ")
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
