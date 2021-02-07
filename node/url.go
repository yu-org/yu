package node

import (
	"net/http"
	"path/filepath"
	. "yu/common"
)

const (
	DownloadUpdatedPath     = "/download/upgrade"
	RegisterNodeKeepersPath = "/nodekeeper/register"
	RegisterWorkersPath     = "/worker/register"
	HeartbeatPath           = "/heartbeat"

	// For developers, every customized Execution and Query of tripods
	// will base on '/api'.
	RootApiPath = "/api"

	TripodNameKey = "tripod"
	CallNameKey   = "call"
	AccountIdKey  = "accountId"
)

var (
	ExecApiPath = filepath.Join(RootApiPath, ExecCallType)
	QryApiPath  = filepath.Join(RootApiPath, QryCallType)
)

// return (Tripod Name, Execution/Query Name)
func GetTripodCallName(req *http.Request) (string, string) {
	query := req.URL.Query()
	return query.Get(TripodNameKey), query.Get(CallNameKey)
}

// return the AccountId of Txn-Sender
func GetAccountId(req *http.Request) Hash {
	return StrToHash(req.URL.Query().Get(AccountIdKey))
}
