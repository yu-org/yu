package poa

import (
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/core/keypair"
)

type PoaConfig struct {
	KeyType string `toml:"key_type"`
	// secret for generating keypair.
	MySecret   string           `toml:"my_secret"`
	Validators []*ValidatorConf `toml:"validators"`
}

type ValidatorConf struct {
	Pubkey string `toml:"pubkey"`
	P2pIp  string `toml:"p2p_ip"`
}

func resolveConfig(cfg *PoaConfig) (PubKey, PrivKey, []ValidatorInfo, error) {
	pub, priv, err := GenKeyPairWithSecret(cfg.KeyType, []byte(cfg.MySecret))
	if err != nil {
		return nil, nil, nil, err
	}
	infos := make([]ValidatorInfo, 0)
	for _, validator := range cfg.Validators {
		pubkey, err := PubkeyFromStr(validator.Pubkey)
		if err != nil {
			return nil, nil, nil, err
		}
		if validator.P2pIp == "" {
			infos = append(infos, ValidatorInfo{
				Pubkey: pubkey,
			})
		} else {
			peerID, err := peer.Decode(validator.P2pIp)
			if err != nil {
				return nil, nil, nil, err
			}
			infos = append(infos, ValidatorInfo{
				Pubkey: pubkey,
				P2pID:  peerID,
			})
		}
	}
	return pub, priv, infos, nil
}
