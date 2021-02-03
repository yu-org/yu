package node

import (
	"encoding/json"
	"reflect"
)

const (
	DownloadUpdatedPath     = "/master/upgrade"
	RegisterNodeKeepersPath = "/nodekeeper/register"
	RegisterWorkersPath     = "/worker/register"
	HeartbeatPath           = "/heartbeat"
)

type WorkerInfo struct {
	Name           string `json:"name"`
	ServesPort     string `json:"serves_port"`
	NodeKeeperAddr string `json:"node_keeper_addr"`
	// Key: Tripod Name
	// Value: Executions Name
	TripodsInfo map[string][]string `json:"tripods_info"`
	Online      bool                `json:"online"`
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
	// key: Worker's Addr
	WorkersInfo map[string]WorkerInfo `json:"workers_info"`
	ServesPort  string                `json:"serves_port"`
	Online      bool                  `json:"online"`
}

func (nki NodeKeeperInfo) Equals(other NodeKeeperInfo) bool {
	return nki.OsArch == other.OsArch && reflect.DeepEqual(nki.WorkersInfo, other.WorkersInfo)
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
