package worker

import (
	"bytes"
	"net/http"
	. "yu/blockchain"
	. "yu/common"
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
	"yu/tripod"
	"yu/txn"
	. "yu/txpool"
)

type Worker struct {
	Name           string
	httpPort       string
	wsPort         string
	NodeKeeperAddr string
	chain          IBlockChain
	txPool         *TxPool
	land           *tripod.Land
	metadb         kv.KV
}

func NewWorker(cfg *config.WorkerConf, chain IBlockChain, txPool *TxPool, land *tripod.Land) (*Worker, error) {
	metadb, err := kv.NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	nkAddr := "localhost:" + cfg.NodeKeeperPort
	return &Worker{
		Name:           cfg.Name,
		httpPort:       ":" + cfg.HttpPort,
		wsPort:         ":" + cfg.WsPort,
		NodeKeeperAddr: nkAddr,
		chain:          chain,
		txPool:         txPool,
		land:           land,
		metadb:         metadb,
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
	for triName, tri := range w.land.Tripods {
		execNames := tri.TripodMeta().AllExecNames()
		queryNames := tri.TripodMeta().AllQueryNames()
		tripodsInfo[triName] = TripodInfo{
			ExecNames:  execNames,
			QueryNames: queryNames,
		}
	}
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

func (w *Worker) putTxpool(req *http.Request, params JsonString) error {
	tripodName, execName := GetTripodCallName(req)
	ecall := &Ecall{
		TripodName: tripodName,
		ExecName:   execName,
		Params:     params,
	}
	caller := GetAddress(req)
	utxn, err := txn.NewUnsignedTxn(caller, ecall)
	if err != nil {
		return err
	}
	stxn, err := utxn.ToSignedTxn()
	if err != nil {
		return err
	}
	return w.txPool.Pend(stxn)
}

func (w *Worker) doQryCall(req *http.Request, params JsonString) error {
	tripodName, qryName := GetTripodCallName(req)
	blockNum, err := GetBlockNumber(req)
	if err != nil {
		return err
	}
	qcall := &Qcall{
		TripodName:  tripodName,
		QueryName:   qryName,
		Params:      params,
		BlockNumber: blockNum,
	}
	return w.land.Query(qcall)
}

func (w *Worker) Run() {

}
