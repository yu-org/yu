package kernel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/yu-org/yu/config"
	"github.com/yu-org/yu/core/protocol"
)

// getChainSpec calls the chain-spec API of a kernel holding the given spec.
func getChainSpec(t *testing.T, spec config.ChainSpec) (int, protocol.APIResponse, config.ChainSpec) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	k := &Kernel{cfg: &config.KernelConf{ChainSpec: spec}}
	r := gin.New()
	r.GET(protocol.ChainSpecPath, k.GetChainSpec)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, protocol.ChainSpecPath, nil))

	var resp struct {
		protocol.APIResponse
		Data config.ChainSpec `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response(%s) error: %v", w.Body.String(), err)
	}
	return w.Code, resp.APIResponse, resp.Data
}

func TestGetChainSpec(t *testing.T) {
	spec := config.ChainSpec{
		ChainName: "Yu",
		Author:    "yu-org",
		Version:   "v1.0.0",
		Network:   config.Mainnet,
	}
	code, resp, got := getChainSpec(t, spec)

	if code != http.StatusOK {
		t.Fatalf("http status = %d, want %d", code, http.StatusOK)
	}
	if !resp.IsSuccess() {
		t.Fatalf("api code = %d, err_msg = %s", resp.Code, resp.ErrMsg)
	}
	if got != spec {
		t.Fatalf("chain-spec = %+v, want %+v", got, spec)
	}
}

func TestGetChainSpecFillsDefaults(t *testing.T) {
	_, _, got := getChainSpec(t, config.ChainSpec{ChainName: "Yu"})

	def := config.DefaultChainSpec()
	if got.Author != def.Author || got.Version != def.Version || got.Network != def.Network {
		t.Fatalf("empty fields not filled with defaults: %+v", got)
	}
	if got.ChainName != "Yu" {
		t.Fatalf("chain_name = %s, want Yu", got.ChainName)
	}
}
