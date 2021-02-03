package worker

import (
	"bytes"
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

// Register into NodeKeeper
func (w *Worker) RegisterInNk() error {
	infoByt, err := w.info.EncodeMasterInfo()
	if err != nil {
		return err
	}
	_, err = w.postToNk(RegisterWorkersPath, infoByt)
	return err
}

func (w *Worker) postToNk(path string, body []byte) (*http.Response, error) {
	url := w.info.NodeKeeperAddr + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	return cli.Do(req)
}
