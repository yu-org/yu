package worker

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
)

type Worker struct {
	info   *WorkerInfo
	metadb kv.KV
}

func NewWorker(cfg *config.WorkerConf) (*Worker, error) {
	metadb, err := kv.NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	nkAddr := "localhost:" + cfg.NodeKeeperPort
	info := &WorkerInfo{
		Name:           cfg.Name,
		Port:           ":" + cfg.ServesPort,
		NodeKeeperAddr: nkAddr,
		Online:         true,
	}
	return &Worker{
		info:   info,
		metadb: metadb,
	}, nil

}

func (w *Worker) HandleHttp() {
	r := gin.Default()

	r.GET(HeartbeatPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})

	r.Run(w.info.Port)
}
