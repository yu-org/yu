package master

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
	"time"
	. "yu/blockchain"
	. "yu/common"
	. "yu/config"
	. "yu/node"
	"yu/storage/kv"
	"yu/tripod"
	. "yu/txpool"
	. "yu/utils/ip"
	. "yu/yerror"
)

type Master struct {
	sync.Mutex
	p2pHost host.Host
	RunMode RunMode
	// Key: NodeKeeper IP, Value: NodeKeeperInfo
	nkDB     kv.KV
	httpPort string
	wsPort   string
	ctx      context.Context
	timeout  time.Duration

	chain  IBlockChain
	txPool ItxPool
	land   *tripod.Land

	// blocks to broadcast into P2P network
	blocksBcChan chan TransferBody
	// txns to broadcast into P2P network
	txnsBcChan chan TransferBody
	// blocks from P2P network
	blocksFromNetChan chan IBlock
}

func NewMaster(cfg *MasterConf, chain IBlockChain, txPool ItxPool, land *tripod.Land) (*Master, error) {
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
		p2pHost:           p2pHost,
		RunMode:           cfg.RunMode,
		nkDB:              nkDB,
		timeout:           timeout,
		ctx:               ctx,
		httpPort:          MakePort(cfg.HttpPort),
		wsPort:            MakePort(cfg.WsPort),
		chain:             chain,
		txPool:            txPool,
		land:              land,
		blocksBcChan:      make(chan TransferBody),
		txnsBcChan:        make(chan TransferBody),
		blocksFromNetChan: make(chan IBlock),
	}, nil
}

func (m *Master) P2pID() string {
	return m.p2pHost.ID().String()
}

// Check the health of NodeKeepers by SendHeartbeat to them.
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
	nkIP := MakeIp(c.ClientIP(), nkInfo.ServesPort)

	err = m.SetNodeKeeper(nkIP, nkInfo)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("store NodeKeeper(%s) error: %s", nkIP, err.Error()),
		)
		return
	}

	c.String(http.StatusOK, "")
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
	err := m.allNodeKeepers(func(ip string, _ *NodeKeeperInfo) error {
		nkIPs = append(nkIPs, ip)
		return nil
	})
	return nkIPs, err
}

func (m *Master) WorkersCount() (int, error) {
	count := 0
	err := m.allNodeKeepers(func(_ string, info *NodeKeeperInfo) error {
		count += len(info.WorkersInfo)
		return nil
	})
	return count, err
}

// find workerIP by Execution/Query name
func (m *Master) findWorkerIP(tripodName, callName string, callType CallType) (ip string, err error) {
	err = m.allNodeKeepers(func(ip string, info *NodeKeeperInfo) error {
		if !info.Online {
			return NodeKeeperDead(ip)
		}
		for workerIp, workerInfo := range info.WorkersInfo {
			if !workerInfo.Online {
				return WorkerDead(workerInfo.Name)
			}
			triInfo, ok := workerInfo.TripodsInfo[tripodName]
			if !ok {
				return TripodNotFound(tripodName)
			}
			var callArr []string
			switch callType {
			case ExecCall:
				callArr = triInfo.ExecNames
			case QryCall:
				callArr = triInfo.QueryNames
			}
			for _, call := range callArr {
				if call == callName {
					ip = workerIp
					return nil
				}
			}
		}
		return nil
	})
	if err != nil {
		return
	}
	if ip == "" {
		switch callType {
		case ExecCall:
			err = ExecNotFound(callName)
		case QryCall:
			err = QryNotFound(callName)
		}
	}
	return
}

func (m *Master) allNodeKeepers(fn func(ip string, info *NodeKeeperInfo) error) error {
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
		err = fn(ip, info)
		if err != nil {
			return err
		}
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
