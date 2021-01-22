package node

import (
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
)

var WorkerNodeInfoKey = []byte("worker-node-info")

type WorkerNode struct {
	info   *WorkerNodeInfo
	metadb kv.KV
}

func NewWorkerNode(cfg *config.Conf) (*WorkerNode, error) {
	metadb, err := kv.NewKV(&cfg.NodeDB)
	if err != nil {
		return nil, err
	}
	data, err := metadb.Get(WorkerNodeInfoKey)
	if err != nil {
		return nil, err
	}
	var info *WorkerNodeInfo
	if data == nil {
		info = &WorkerNodeInfo{
			Name:        cfg.NodeConf.NodeName,
			MasterNodes: cfg.NodeConf.MasterNodes,
		}
		infoByt, err := info.EncodeMasterNodeInfo()
		if err != nil {
			return nil, err
		}
		err = metadb.Set(WorkerNodeInfoKey, infoByt)
		if err != nil {
			return nil, err
		}
	} else {
		info, err = DecodeWorkerNodeInfo(data)
		if err != nil {
			return nil, err
		}
	}

	return &WorkerNode{
		info,
		metadb,
	}, nil

}
