package MEVless

type Config struct {
	PackNumber uint64 `toml:"pack_number"`
	Addr       string `toml:"addr"`
	Charge     uint64 `toml:"charge"`
	DbPath     string `toml:"db_path"`
}

func DefaultCfg() *Config {
	return &Config{
		PackNumber: 10000,
		Addr:       "localhost:9071",
		Charge:     1000,
		DbPath:     "yu/mev_less",
	}
}
