package node

import (
	"encoding/json"
)

type MasterInfo struct {
	P2pID       string   `json:"p2p_id"`
	Name        string   `json:"name"`
	WorkerNodes []string `json:"worker_nodes"`
}

func (mi *MasterInfo) EncodeMasterInfo() ([]byte, error) {
	return json.Marshal(mi)
}

func DecodeMasterInfo(data []byte) (*MasterInfo, error) {
	var info MasterInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

type WorkerInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	MasterNode string `json:"master_node"`
}

func (wi *WorkerInfo) EncodeMasterInfo() ([]byte, error) {
	return json.Marshal(wi)
}

func DecodeWorkerInfo(data []byte) (*WorkerInfo, error) {
	var info WorkerInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
