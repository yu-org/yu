package worker

import (
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
)

var WorkerInfoKey = []byte("worker-node-info")

type Worker struct {
	info   *WorkerInfo
	metadb kv.KV
}

func NewWorker(cfg *config.WorkerConf) (*Worker, error) {
	metadb, err := kv.NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	data, err := metadb.Get(WorkerInfoKey)
	if err != nil {
		return nil, err
	}
	var info *WorkerInfo
	if data == nil {
		info = &WorkerInfo{
			Name:           cfg.Name,
			NodeKeeperAddr: cfg.NodeKeeperAddr,
			Port:           ":" + cfg.ServesPort,
		}
		infoByt, err := info.EncodeMasterInfo()
		if err != nil {
			return nil, err
		}
		err = metadb.Set(WorkerInfoKey, infoByt)
		if err != nil {
			return nil, err
		}
	} else {
		info, err = DecodeWorkerInfo(data)
		if err != nil {
			return nil, err
		}
	}

	return &Worker{
		info,
		metadb,
	}, nil

}
