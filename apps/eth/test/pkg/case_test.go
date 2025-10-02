package pkg

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/os"
)

func TestTransferStep(t *testing.T) {
	data, err := os.ReadFile("./test/data.json")
	require.NoError(t, err)
	tc := &TransferCase{}
	require.NoError(t, json.Unmarshal(data, tc))
	require.Equal(t, calculateSteps(tc.Steps, copyOrigin(tc.Original)), tc.Expect)
	ntc := &TransferCase{
		Steps:    tc.Steps,
		Original: tc.Original,
		Expect:   tc.Original,
	}
	calculateExpect(ntc)
	require.Equal(t, tc.Expect, ntc.Expect)
}

func copyOrigin(origin map[string]*CaseEthWallet) map[string]*CaseEthWallet {
	m := make(map[string]*CaseEthWallet)
	for k, v := range origin {
		m[k] = v.Copy()
	}
	return m
}

func calculateSteps(steps []*Step, origin map[string]*CaseEthWallet) map[string]*CaseEthWallet {
	for _, step := range steps {
		calculateStep(step, origin)
	}
	return origin
}

func calculateStep(step *Step, origin map[string]*CaseEthWallet) {
	origin[step.From.Address].EthCount = origin[step.From.Address].EthCount - step.Count
	origin[step.To.Address].EthCount = origin[step.To.Address].EthCount + step.Count
}
