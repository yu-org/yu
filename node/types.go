package node

import "encoding/json"

type NodeType = uint

const (
	Master NodeType = iota
	Worker
)

type MasterNodeInfo struct {
	P2pID       string   `json:"p2p_id"`
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	MasterNodes []string `json:"master_nodes"`
	WorkerNodes []string `json:"worker_nodes"`
}

func (mi *MasterNodeInfo) EncodeMasterNodeInfo() ([]byte, error) {
	return json.Marshal(mi)
}

func DecodeMasterNodeInfo(data []byte) (*MasterNodeInfo, error) {
	var info MasterNodeInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

type WorkerNodeInfo struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	MasterNodes []string `json:"master_nodes"`
}

func (wi *WorkerNodeInfo) EncodeMasterNodeInfo() ([]byte, error) {
	return json.Marshal(wi)
}

func DecodeWorkerNodeInfo(data []byte) (*WorkerNodeInfo, error) {
	var info WorkerNodeInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
