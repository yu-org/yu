package node

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func SendHeartbeats(addrs []string) {
	for _, addr := range addrs {
		_, err := http.Get(addr + HeartbeatToPath)
		if err != nil {
			logrus.Errorf("send heartbeat to (%s) error: %s", addr, err.Error())
		} else {
			logrus.Debugf("send heartbeat to (%s) succeed!", addr)
		}
	}

}

func ReplyHeartbeat(e *gin.Engine) {
	e.GET(HeartbeatToPath, func(c *gin.Context) {
		c.String(http.StatusOK, "")
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})
}
