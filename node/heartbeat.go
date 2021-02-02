package node

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
)

func SendHeartbeats(addrs []string, handleDead func(addr string) error) {
	for _, addr := range addrs {
		_, err := http.Get(addr + HeartbeatToPath)
		if err != nil {
			logrus.Errorf("send heartbeat to (%s) error: %s", addr, err.Error())
			err = handleDead(addr)
			if err != nil {
				logrus.Errorf("handle dead node (%s) error: %s", addr, err.Error())
			}
		} else {
			logrus.Debugf("send heartbeat to (%s) succeed!", addr)
		}
	}

}

func ReplyHeartbeat(e *gin.Engine, replyBody interface{}) {
	e.GET(HeartbeatToPath, func(c *gin.Context) {
		c.String(http.StatusOK, "")
		c.JSON(http.StatusOK, replyBody)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})
}
