package node

import (
	"github.com/gin-gonic/gin"
	"path/filepath"
	"strings"
)

const (
	DownloadUpdatedPath     = "/download/upgrade"
	RegisterNodeKeepersPath = "/nodekeeper/register"
	RegisterWorkersPath     = "/worker/register"
	HeartbeatPath           = "/heartbeat"

	// For developers, every customized Execution and Query of tripods
	// will base on '/api'.
	RootApiPath = "/api"
	// Use to match every api-path of Tripods' Name and their Execution/Query Name
	WildcardApiPath = "*api"

	ExecCallType = "execution"
	QryCallType  = "query"
)

var (
	ExecApiPath = filepath.Join(RootApiPath, ExecCallType, WildcardApiPath)
	QryApiPath  = filepath.Join(RootApiPath, QryCallType, WildcardApiPath)
)

// return (Tripod Name, Execution/Query Name)
func ResolveApiUrl(c *gin.Context) (string, string) {
	apiUrl := c.Param("api")
	names := strings.Split(apiUrl, "/")
	return names[1], names[2]
}
