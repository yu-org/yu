package node

import (
	"encoding/json"
	"reflect"
)

const (
	DownloadUpdatedPath  = "/master/upgrade"
	WatchNodeKeepersPath = "/nodekeeper"
	WatchWorkersPath     = "/worker"
)

type MasterInfo struct {
	P2pID string `json:"p2p_id"`
	// key: NodeKeeper's IP
	NodeKeepersInfo map[string]NodeKeeperInfo `json:"node_keepers_info"`
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
	Port           string `json:"port"`
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
	OsArch string `json:"os_arch"`
	// key: Worker's ID
	WorkersStatus map[int]WorkerStatus `json:"workers_status"`
}

func (nki NodeKeeperInfo) Equals(other NodeKeeperInfo) bool {
	return nki.OsArch == other.OsArch && reflect.DeepEqual(nki.WorkersStatus, other.WorkersStatus)
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

type WorkerStatus struct {
	Info   WorkerInfo `json:"info"`
	Online bool       `json:"online"`
}

func (ws *WorkerStatus) EncodeWorkerStatus() ([]byte, error) {
	return json.Marshal(ws)
}

func DecodeWorkerStatus(data []byte) (*WorkerStatus, error) {
	var info WorkerStatus
	err := json.Unmarshal(data, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}
