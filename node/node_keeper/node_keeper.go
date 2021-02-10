package node_keeper

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
	"yu/config"
	. "yu/node"
	"yu/storage/kv"
	"yu/utils/compress"
)

const CompressedFileType = ".zip"

type NodeKeeper struct {
	// Key: repoName,  Value: Repo
	repoDB kv.KV
	// Workers: Key worker_addr, Value workerInfo
	workerDB   kv.KV
	dir        string
	servesPort string
	masterAddr string

	osArch  string
	timeout time.Duration
}

func NewNodeKeeper(cfg *config.NodeKeeperConf) (*NodeKeeper, error) {
	dir := cfg.Dir
	repoDbPath := compactPath(dir, cfg.RepoDbPath)
	workerDbPath := compactPath(dir, cfg.WorkerDbPath)

	repoDB, err := kv.NewBolt(repoDbPath)
	if err != nil {
		return nil, err
	}
	workerDB, err := kv.NewBolt(workerDbPath)
	if err != nil {
		return nil, err
	}

	// get OS and Arch from local host
	var osArch string
	if cfg.OsArch == "" {
		osArch = runtime.GOOS + "-" + runtime.GOARCH
	}

	timeout := time.Duration(cfg.Timeout) * time.Second

	return &NodeKeeper{
		repoDB:     repoDB,
		workerDB:   workerDB,
		dir:        dir,
		servesPort: ":" + cfg.ServesPort,
		masterAddr: cfg.MasterAddr,
		osArch:     osArch,
		timeout:    timeout,
	}, nil
}

// Register master the the existence of Nodekeeper and the changes of workers' number.
// When workers increase or decrease or keep-alive, should POST to master.
func (n *NodeKeeper) RegisterInMaster() error {
	info, err := n.Info()
	if err != nil {
		return err
	}
	byt, err := info.EncodeNodeKeeperInfo()
	if err != nil {
		return err
	}
	_, err = n.postToMaster(RegisterNodeKeepersPath, byt)
	return err
}

func (n *NodeKeeper) HandleHttp() {
	r := gin.Default()

	// Handle from Master. When upgrade onchain, Master will give out
	// updated executable compressed package to each worker.
	r.POST(DownloadUpdatedPath, func(c *gin.Context) {
		n.downloadUpdatedPkg(c)
		logrus.Info("download updated package succeed")
	})

	// Handle from worker. Used for watch the changes of workers
	// and report to Master.
	r.POST(RegisterWorkersPath, func(c *gin.Context) {
		n.registerWorkers(c)
	})

	r.GET(HeartbeatPath, func(c *gin.Context) {
		info, err := n.Info()
		if err != nil {
			logrus.Errorf("get NodeKeeper info error: %s", err.Error())
			c.JSON(http.StatusInternalServerError, nil)
			return
		}
		c.JSON(http.StatusOK, info)
		logrus.Debugf("accept heartbeat from %s", c.ClientIP())
	})

	r.Run(n.servesPort)
}

// Check the health of Workers by SendHeartbeat to them.
func (n *NodeKeeper) CheckHealth() {
	for {
		wAddrs, err := n.allWorkersIP()
		if err != nil {
			logrus.Errorf("get all Workers error: %s", err.Error())
		}
		SendHeartbeats(wAddrs, func(addr string) error {
			tx, err := n.workerDB.NewKvTxn()
			if err != nil {
				return err
			}
			workerInfo, err := getWorkerWithTx(tx, addr)
			if err != nil {
				return err
			}
			workerInfo.Online = false
			err = setWorkerWithTx(tx, addr, workerInfo)
			if err != nil {
				return err
			}
			return tx.Commit()
		})
		time.Sleep(n.timeout)
	}
}

// Watch the changes of workers' number.
// When workers increase or decrease, should request this API.
func (n *NodeKeeper) registerWorkers(c *gin.Context) {
	var workerInfo WorkerInfo
	err := c.ShouldBindJSON(&workerInfo)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("bad worker-info data struct: %s", err.Error()),
		)
		return
	}
	workerAddr := c.ClientIP() + workerInfo.HttpPort
	err = n.setWorkerInfo(workerAddr, &workerInfo)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("set new worker(%s) error: %s", workerAddr, err.Error()),
		)
		return
	}

	logrus.Infof("Register Worker(%v) into NodeKeeper succeed. ", workerAddr)

	err = n.RegisterInMaster()
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("nortify master error: %s", err.Error()),
		)
		return
	}
	c.String(http.StatusOK, "")
	logrus.Infof("Register Worker(%v) into Master succeed. ", workerAddr)
}

func (n *NodeKeeper) downloadUpdatedPkg(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("upload file error: %s", err.Error()))
		return
	}

	fname := file.Filename
	if !strings.HasSuffix(fname, CompressedFileType) {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("the type of file(%s) is wrong", fname),
		)
		return
	}
	zipFileName := filepath.Join(n.dir, fname)
	err = c.SaveUploadedFile(file, zipFileName)
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("save file(%s) error: %s", fname, err.Error()),
		)
		return
	}
	err = n.convertToRepo(zipFileName, fname)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("convert file(%s) to repo error: %s", fname, err.Error()),
		)
	}

	c.String(http.StatusOK, "download file succeed")
}

// zipFilePath just like: path/to/yuRepo_linux-amd64_3.zip
// 'yuRepo' is the name of repo
// 'linux-amd64' is the arch of repo
// '3' is the version of repo
func (n *NodeKeeper) convertToRepo(zipFilePath, fname string) error {

	// repoVDir: path/to/yuRepo_linux-amd64_3
	repoVDir := strings.TrimSuffix(zipFilePath, CompressedFileType)

	arr := strings.Split(repoVDir, "_")
	repoVersionStr := arr[len(arr)-1]
	repoVersion, err := strconv.Atoi(repoVersionStr)
	if err != nil {
		return err
	}

	// repoArch: linux-amd64
	repoArch := arr[len(arr)-2]

	// repoName: yuRepo
	repoName := strings.TrimSuffix(fname, "_"+repoArch+"_"+repoVersionStr+CompressedFileType)

	// repoDir: path/to/yuRepo/3/linux-amd64
	repoDir := filepath.Join(n.dir, repoName, repoVersionStr, repoArch)
	err = os.MkdirAll(repoDir, os.ModePerm)
	if err != nil {
		return err
	}

	files, err := compress.UnzipFile(zipFilePath, repoDir)
	if err != nil {
		return err
	}

	repo, err := NewRepo(repoName, files, repoDir, repoVersion)
	if err != nil {
		return err
	}
	err = n.setRepo(repoName, repo)
	if err != nil {
		return err
	}
	return os.Remove(zipFilePath)
}

func (n *NodeKeeper) postToMaster(path string, body []byte) (*http.Response, error) {
	url := n.masterAddr + path
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	return cli.Do(req)
}

func (n *NodeKeeper) getWorkerInfo(addr string) (*WorkerInfo, error) {
	infoByt, err := n.workerDB.Get([]byte(addr))
	if err != nil {
		return nil, err
	}
	return DecodeWorkerInfo(infoByt)
}

func (n *NodeKeeper) setWorkerInfo(addr string, info *WorkerInfo) error {
	infoByt, err := info.EncodeMasterInfo()
	if err != nil {
		return err
	}
	return n.workerDB.Set([]byte(addr), infoByt)
}

func (n *NodeKeeper) getRepo(repoName string) (*Repo, error) {
	repoByt, err := n.repoDB.Get([]byte(repoName))
	if err != nil {
		return nil, err
	}
	return decodeRepo(repoByt)
}

func (n *NodeKeeper) setRepo(repoName string, repo *Repo) error {
	repoByt, err := repo.encode()
	if err != nil {
		return err
	}
	return n.repoDB.Set([]byte(repoName), repoByt)
}

func (n *NodeKeeper) Info() (*NodeKeeperInfo, error) {
	workersInfo := make(map[string]WorkerInfo)
	err := n.allWorkers(func(addr string, info *WorkerInfo) {
		workersInfo[addr] = *info
	})
	if err != nil {
		return nil, err
	}
	return &NodeKeeperInfo{
		OsArch:      n.osArch,
		WorkersInfo: workersInfo,
		ServesPort:  n.servesPort,
		Online:      true,
	}, nil
}

func (n *NodeKeeper) allWorkersIP() ([]string, error) {
	workersIP := make([]string, 0)
	err := n.allWorkers(func(addr string, _ *WorkerInfo) {
		workersIP = append(workersIP, addr)
	})
	return workersIP, err
}

func (n *NodeKeeper) allWorkers(fn func(addr string, info *WorkerInfo)) error {
	iter, err := n.workerDB.Iter(nil)
	if err != nil {
		return err
	}
	defer iter.Close()
	for iter.Valid() {
		addrByt, infoByt, err := iter.Entry()
		if err != nil {
			return err
		}
		addr := string(addrByt)
		info, err := DecodeWorkerInfo(infoByt)
		if err != nil {
			return err
		}
		fn(addr, info)
		err = iter.Next()
		if err != nil {
			return err
		}
	}
	return nil
}

func getWorkerWithTx(tx kv.KvTxn, ip string) (*WorkerInfo, error) {
	infoByt, err := tx.Get([]byte(ip))
	if err != nil {
		return nil, err
	}
	return DecodeWorkerInfo(infoByt)
}

func setWorkerWithTx(tx kv.KvTxn, ip string, info *WorkerInfo) error {
	infoByt, err := info.EncodeMasterInfo()
	if err != nil {
		return err
	}
	return tx.Set([]byte(ip), infoByt)
}

func compactPath(dir, path string) string {
	if filepath.Dir(path) == dir {
		path = filepath.Join(dir, filepath.Base(path))
	}
	return path
}
