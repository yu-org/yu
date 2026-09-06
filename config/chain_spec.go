package config

// Network is the category of the network which the chain is running on,
// such as "mainnet", "testnet" ...
type Network = string

const (
	Mainnet Network = "mainnet"
	Testnet Network = "testnet"
	Devnet  Network = "devnet"
)

// ChainSpec describes the identity of the chain.
// It does NOT affect the consensus, it is only the self-description of this chain,
// printed out at the very beginning of running.
type ChainSpec struct {
	// Name of the chain, such as "Yu".
	ChainName string `toml:"chain_name"`
	// Who maintains this chain.
	Author string `toml:"author"`
	// Version of this chain, such as "v1.0.0".
	Version string `toml:"version"`
	// Which kind of network this chain is running on:
	// mainnet, testnet, devnet.
	Network Network `toml:"network"`
}

func DefaultChainSpec() ChainSpec {
	return ChainSpec{
		ChainName: "Yu",
		Author:    "yu-org",
		Version:   "v1.0.0",
		Network:   Devnet,
	}
}

// FillDefaults fills the empty fields with the default chain-spec,
// so that a config-file without [chain_spec] still prints a meaningful banner.
func (c *ChainSpec) FillDefaults() {
	def := DefaultChainSpec()
	if c.ChainName == "" {
		c.ChainName = def.ChainName
	}
	if c.Author == "" {
		c.Author = def.Author
	}
	if c.Version == "" {
		c.Version = def.Version
	}
	if c.Network == "" {
		c.Network = def.Network
	}
}
