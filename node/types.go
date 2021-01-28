package node

import (
	"encoding/json"
)

type MasterInfo struct {
	P2pID        string   `json:"p2p_id"`
	Name         string   `json:"name"`
	WorkersAddrs []string `json:"workers_addrs"`
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
	ID             int    `json:"id"`
	Name           string `json:"name"`
	NodeKeeperAddr string `json:"node_keeper_addr"`
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

type NodeKeeperInfo struct {
	OsArch      string       `json:"os_arch"`
	WorkersInfo []WorkerInfo `json:"workers_info"`
}

func (nki *NodeKeeperInfo) EncodeNodeKeeperInfo() ([]byte, error) {
	return json.Marshal(nki)
}

func DecodeNodeKeeperInfo(data []byte) (*NodeKeeperInfo, error) {
	var info NodeKeeperInfo
	err := json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
