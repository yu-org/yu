package banner

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/yu-org/yu/config"
)

func TestFprintChainSpec(t *testing.T) {
	var buf bytes.Buffer
	FprintChainSpec(&buf, &config.ChainSpec{
		ChainName: "Yu",
		Author:    "yu-org",
		Version:   "v1.0.0",
		Network:   config.Mainnet,
	}, false)

	out := buf.String()
	for _, want := range []string{"yu-org", "v1.0.0", "MAINNET", "█"} {
		if !strings.Contains(out, want) {
			t.Fatalf("banner missing %q, got:\n%s", want, out)
		}
	}
	if strings.Contains(out, "\033[") {
		t.Fatalf("no-color banner contains ANSI codes:\n%s", out)
	}
}

func TestFprintChainSpecFillsDefaults(t *testing.T) {
	spec := new(config.ChainSpec)
	var buf bytes.Buffer
	FprintChainSpec(&buf, spec, false)

	def := config.DefaultChainSpec()
	if !strings.Contains(buf.String(), def.Author) {
		t.Fatalf("empty spec did not fall back to defaults, got:\n%s", buf.String())
	}
	// the caller's config stays untouched
	if spec.Author != "" {
		t.Fatal("FprintChainSpec must not modify the given ChainSpec")
	}
}

// TestShowChainSpec is for eyeballing the banner: go test ./utils/banner -run ShowChainSpec -v
func TestShowChainSpec(t *testing.T) {
	for _, network := range []config.Network{config.Mainnet, config.Testnet, config.Devnet} {
		FprintChainSpec(os.Stdout, &config.ChainSpec{
			ChainName: "Yu Chain",
			Author:    "yu-org",
			Version:   "v1.0.0",
			Network:   network,
		}, true)
	}
}
