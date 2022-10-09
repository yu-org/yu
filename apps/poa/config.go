package poa

import (
	"github.com/libp2p/go-libp2p-core/peer"
	. "github.com/yu-org/yu/core/keypair"
)

type PoaConfig struct {
	KeyType string `json:"key_type"`
	// secret for generating keypair.
	MySecret   string           `json:"my_secret"`
	Validators []*ValidatorConf `json:"validators"`
}

type ValidatorConf struct {
	Pubkey string `json:"pubkey"`
	P2pIp  string `json:"p2p_ip"`
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
		peerID, err := peer.Decode(validator.P2pIp)
		if err != nil {
			return nil, nil, nil, err
		}
		infos = append(infos, ValidatorInfo{
			Pubkey: pubkey,
			P2pID:  peerID,
		})
	}
	return pub, priv, infos, nil
}
