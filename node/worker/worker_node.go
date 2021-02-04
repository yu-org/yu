package worker

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
	"yu/tripod"
)

type Worker struct {
	Name           string
	ServesPort     string
	NodeKeeperAddr string
	land           *tripod.Land
	metadb         kv.KV
}

func NewWorker(cfg *config.WorkerConf, land *tripod.Land) (*Worker, error) {
	metadb, err := kv.NewKV(&cfg.DB)
	if err != nil {
		return nil, err
	}
	nkAddr := "localhost:" + cfg.NodeKeeperPort
	return &Worker{
		Name:           cfg.Name,
		ServesPort:     ":" + cfg.ServesPort,
		NodeKeeperAddr: nkAddr,
		land:           land,
		metadb:         metadb,
	}, nil

}

func (w *Worker) HandleHttp() {
	r := gin.Default()

	r.GET(HeartbeatPath, func(c *gin.Context) {
		c.JSON(http.StatusOK, nil)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})

	r.POST(ExecApiPath, func(c *gin.Context) {
		tripodName, execName := ResolveApiUrl(c)
	})

	r.POST(QryApiPath, func(c *gin.Context) {
		tripodName, qryName := ResolveApiUrl(c)
	})

	r.Run(w.ServesPort)
}

// Register into NodeKeeper
func (w *Worker) RegisterInNk() error {
	infoByt, err := w.Info().EncodeMasterInfo()
	if err != nil {
		return err
	}
	_, err = w.postToNk(RegisterWorkersPath, infoByt)
	return err
}

func (w *Worker) Info() *WorkerInfo {
	tripodsInfo := make(map[string]TripodInfo)
	for triName, tri := range w.land.Tripods {
		execNames := tri.TripodMeta().AllExecNames()
		queryNames := tri.TripodMeta().AllQueryNames()
		tripodsInfo[triName] = TripodInfo{
			ExecNames:  execNames,
			QueryNames: queryNames,
		}
	}
	return &WorkerInfo{
		Name:           w.Name,
		ServesPort:     w.ServesPort,
		NodeKeeperAddr: w.NodeKeeperAddr,
		TripodsInfo:    tripodsInfo,
		Online:         true,
	}
}

func (w *Worker) postToNk(path string, body []byte) (*http.Response, error) {
	url := w.NodeKeeperAddr + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	return cli.Do(req)
}
