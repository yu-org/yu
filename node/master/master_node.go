package master

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/host"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/protocol"
	maddr "github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"sync"
	"time"
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
)

type Master struct {
	sync.Mutex
	p2pHost host.Host
	// Key: NodeKeeper IP, Value: NodeKeeperInfo
	nkDB    kv.KV
	port    string
	ctx     context.Context
	timeout time.Duration
}

func NewMaster(cfg *config.MasterConf) (*Master, error) {
	nkDB, err := kv.NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	p2pHost, err := makeP2pHost(ctx, cfg)
	if err != nil {
		return nil, err
	}

	timeout := time.Duration(cfg.Timeout) * time.Second

	return &Master{
		p2pHost: p2pHost,
		nkDB:    nkDB,
		timeout: timeout,
		ctx:     ctx,
		port:    ":" + cfg.ServesPort,
	}, nil
}

func (m *Master) P2pID() string {
	return m.p2pHost.ID().String()
}

func (m *Master) ConnectP2PNetwork(cfg *config.MasterConf) error {
	m.p2pHost.SetStreamHandler(protocol.ID(cfg.ProtocolID), m.handleStream)

	for _, addrStr := range cfg.ConnectAddrs {
		addr, err := maddr.NewMultiaddr(addrStr)
		if err != nil {
			return err
		}
		peer, err := peerstore.AddrInfoFromP2pAddr(addr)
		if err != nil {
			return err
		}
		err = m.p2pHost.Connect(m.ctx, *peer)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *Master) HandleHttp() {
	r := gin.Default()

	r.POST(RegisterNodeKeepersPath, func(c *gin.Context) {
		m.registerNodeKeepers(c)
	})

	r.Run(m.port)
}

// check the health of NodeKeepers by SendHeartbeat to them.
func (m *Master) CheckHealth() {
	for {
		nkAddrs, err := m.allNodeKeepersIp()
		if err != nil {
			logrus.Errorf("get all NodeKeepers error: %s", err.Error())
		}
		SendHeartbeats(nkAddrs, func(nkAddr string) error {
			tx, err := m.nkDB.NewKvTxn()
			if err != nil {
				return err
			}
			info, err := getNkWithTx(tx, nkAddr)
			if err != nil {
				return err
			}
			info.Online = false
			err = setNkWithTx(tx, nkAddr, info)
			if err != nil {
				return err
			}
			return tx.Commit()
		})
		time.Sleep(m.timeout)
	}
}

func (m *Master) registerNodeKeepers(c *gin.Context) {
	m.Lock()
	defer m.Unlock()
	var nkInfo NodeKeeperInfo
	err := c.ShouldBindJSON(&nkInfo)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("NodeKeeperInfo decode failed: %s", err.Error()),
		)
		return
	}
	nkIP := c.ClientIP()

	err = m.SetNodeKeeper(nkIP, nkInfo)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("store NodeKeeper(%s) error: %s", nkIP, err.Error()),
		)
		return
	}

	workerId, err := m.WorkersCount()
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("generate workerID error: %s", err.Error()),
		)
		return
	}

	c.String(http.StatusOK, strconv.Itoa(workerId))
	logrus.Infof("NodeKeeper(%s) register succeed!", nkIP)
}

func (m *Master) SetNodeKeeper(ip string, info NodeKeeperInfo) error {
	infoByt, err := info.EncodeNodeKeeperInfo()
	if err != nil {
		return err
	}
	return m.nkDB.Set([]byte(ip), infoByt)
}

func (m *Master) allNodeKeepersIp() ([]string, error) {
	nkIPs := make([]string, 0)
	err := m.allNodeKeepers(func(ip string, _ *NodeKeeperInfo) {
		nkIPs = append(nkIPs, ip)
	})
	return nkIPs, err
}

func (m *Master) WorkersCount() (int, error) {
	count := 0
	err := m.allNodeKeepers(func(_ string, info *NodeKeeperInfo) {
		count += len(info.WorkersInfo)
	})
	return count, err
}

func (m *Master) allNodeKeepers(fn func(ip string, info *NodeKeeperInfo)) error {
	iter, err := m.nkDB.Iter(nil)
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.Valid() {
		ipByt, infoByt, err := iter.Entry()
		if err != nil {
			return err
		}
		ip := string(ipByt)
		info, err := DecodeNodeKeeperInfo(infoByt)
		if err != nil {
			return err
		}
		fn(ip, info)
		err = iter.Next()
		if err != nil {
			return err
		}
	}
	return nil
}
func getNkWithTx(txn kv.KvTxn, ip string) (*NodeKeeperInfo, error) {
	infoByt, err := txn.Get([]byte(ip))
	if err != nil {
		return nil, err
	}
	return DecodeNodeKeeperInfo(infoByt)
}

func setNkWithTx(txn kv.KvTxn, ip string, info *NodeKeeperInfo) error {
	infoByt, err := info.EncodeNodeKeeperInfo()
	if err != nil {
		return err
	}
	return txn.Set([]byte(ip), infoByt)
}
