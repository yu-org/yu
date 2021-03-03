package worker

import (
	"bytes"
	"net/http"
	. "yu/blockchain"
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
	"yu/tripod"
	"yu/txn"
	. "yu/txpool"
	. "yu/utils/ip"
)

type Worker struct {
	Name           string
	httpPort       string
	wsPort         string
	NodeKeeperAddr string
	chain          IBlockChain
	txPool         ItxPool
	land           *tripod.Land
	metadb         kv.KV

	// ready to package a batch of txns to broadcast
	readyBcTxnsChan chan txn.IsignedTxn
	// number of broadcast txns every time
	NumOfBcTxns int
}

func NewWorker(cfg *config.WorkerConf, chain IBlockChain, txPool ItxPool, land *tripod.Land) (*Worker, error) {
	metadb, err := kv.NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	nkAddr := MakeLocalIp(cfg.NodeKeeperPort)
	return &Worker{
		Name:            cfg.Name,
		httpPort:        MakePort(cfg.HttpPort),
		wsPort:          MakePort(cfg.WsPort),
		NodeKeeperAddr:  nkAddr,
		chain:           chain,
		txPool:          txPool,
		land:            land,
		metadb:          metadb,
		readyBcTxnsChan: make(chan txn.IsignedTxn),
		NumOfBcTxns:     cfg.NumOfBcTxns,
	}, nil

}

// Register into NodeKeeper
func (w *Worker) RegisterInNk() error {
	infoByt, err := w.Info().EncodeMasterInfo()
	if err != nil {
		return err
	}
	_, err = w.postToNk(RegisterWorkersPath, infoByt)
	return err
}

func (w *Worker) Info() *WorkerInfo {
	tripodsInfo := make(map[string]TripodInfo)
	_ = w.land.RangeMap(func(triName string, tri tripod.Tripod) error {
		execNames := tri.TripodMeta().AllExecNames()
		queryNames := tri.TripodMeta().AllQueryNames()
		tripodsInfo[triName] = TripodInfo{
			ExecNames:  execNames,
			QueryNames: queryNames,
		}
		return nil
	})
	return &WorkerInfo{
		Name:           w.Name,
		HttpPort:       w.httpPort,
		WsPort:         w.wsPort,
		NodeKeeperAddr: w.NodeKeeperAddr,
		TripodsInfo:    tripodsInfo,
		Online:         true,
	}
}

func (w *Worker) postToNk(path string, body []byte) (*http.Response, error) {
	url := w.NodeKeeperAddr + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	return cli.Do(req)
}

func (w *Worker) Run() {

}
