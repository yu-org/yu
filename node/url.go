package node

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"path/filepath"
	"strings"
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
	// Use to match every api-path of Tripods' Name and their Execution/Query Name
	WildcardApiPath = "*api"
)

var (
	ExecApiHttpPath = filepath.Join(RootApiPath, ExecCallType, WildcardApiPath)
	QryApiHttpPath  = filepath.Join(RootApiPath, QryCallType, WildcardApiPath)

	ExecApiWsPath = filepath.Join(RootApiPath, ExecCallType)
	QryApiWsPath  = filepath.Join(RootApiPath, QryCallType)
)

// return (Tripod Name, Execution/Query Name)
func ResolveHttpApiUrl(c *gin.Context) (string, string) {
	apiUrl := c.Param("api")
	names := strings.Split(apiUrl, "/")
	return names[1], names[2]
}

// return (Tripod Name, Execution/Query Name)
func ResolveWsApiUrl(req *http.Request) (string, string) {
	urlPath := req.URL.Path
}
