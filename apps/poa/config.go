package poa

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	. "github.com/yu-org/yu/core/keypair"
)

type PoaConfig struct {
	KeyType string `toml:"key_type"`
	// secret for generating keypair.
	MySecret   string           `toml:"my_secret"`
	Validators []*ValidatorConf `toml:"validators"`
	// block out interval, seconds
	BlockInterval int `toml:"block_interval"`
	// the number of packing txns from txpool, default 5000
	PackNum uint64 `toml:"pack_num"`
}

var DefaultSecrets = []string{
	"node1",
	"node2",
	"node3",
}

func DefaultCfg(idx int) *PoaConfig {
	cfg := &PoaConfig{
		KeyType:  Sr25519,
		MySecret: DefaultSecrets[idx],
		Validators: []*ValidatorConf{
			{Pubkey: "", P2pIp: "12D3KooWHHzSeKaY8xuZVzkLbKFfvNgPPeKhFBGrMbNzbm5akpqu"},
			{Pubkey: "", P2pIp: "12D3KooWSKPs95miv8wzj3fa5HkJ1tH7oEGumsEiD92n2MYwRtQG"},
			{Pubkey: "", P2pIp: "12D3KooWRuwP7nXaRhZrmoFJvPPGat2xPafVmGpQpZs5zKMtwqPH"},
		},
		BlockInterval: 3,
		PackNum:       30000,
	}
	var myPubkey PubKey
	for i, secret := range DefaultSecrets {
		pub, _ := GenSrKeyWithSecret([]byte(secret))
		logrus.Infof("pub%d is %s", i, pub.String())
		cfg.Validators[i].Pubkey = pub.StringWithType()
		if idx == i {
			myPubkey = pub
		}
	}
	logrus.Info("My Address is ", myPubkey.Address().String())
	return cfg
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
