package node

import (
	"github.com/gin-gonic/gin"
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
)

var (
	ExecApiPath = filepath.Join(RootApiPath, ExecCallType)
	QryApiPath  = filepath.Join(RootApiPath, QryCallType)
)

// return (Tripod Name, Execution/Query Name)
func ResolveHttpApiUrl(c *gin.Context) (string, string) {
	return c.Query(TripodNameKey), c.Query(CallNameKey)
}

// return (Tripod Name, Execution/Query Name)
func ResolveWsApiUrl(req *http.Request) (string, string) {
	query := req.URL.Query()
	return query.Get(TripodNameKey), query.Get(CallNameKey)
}
