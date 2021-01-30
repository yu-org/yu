package node_keeper

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"yu/config"
	"yu/node"
	"yu/storage/kv"
	"yu/utils/compress"
)

const CompressedFileType = ".zip"

type NodeKeeper struct {
	info       *node.NodeKeeperInfo
	repoDB     kv.KV
	dir        string
	port       string
	masterAddr string
}

func NewNodeKeeper(cfg *config.NodeKeeperConf) (*NodeKeeper, error) {
	dbpath := cfg.RepoDbPath
	dir := cfg.Dir
	if filepath.Dir(dbpath) == dir {
		dbpath = filepath.Join(dir, filepath.Base(dbpath))
	}
	repoDB, err := kv.NewBolt(dbpath)
	if err != nil {
		return nil, err
	}

	// get OS and Arch from local host
	var osArch string
	if cfg.OsArch == "" {
		osArch = runtime.GOOS + "-" + runtime.GOARCH
	}

	info := &node.NodeKeeperInfo{
		OsArch:        osArch,
		WorkersStatus: make(map[int]node.WorkerStatus),
	}

	return &NodeKeeper{
		info:       info,
		repoDB:     repoDB,
		dir:        dir,
		port:       ":" + cfg.ServesPort,
		masterAddr: cfg.MasterAddr,
	}, nil
}

// Nortify master the the existence of Nodekeeper and the changes of workers' number.
// When workers increase or decrease or keep-alive, should POST to master.
func (n *NodeKeeper) NortifyMaster() error {
	infoByt, err := n.info.EncodeNodeKeeperInfo()
	if err != nil {
		return err
	}
	_, err = n.postToMaster("/nodekeeper/worker", infoByt)
	return err
}

func (n *NodeKeeper) HandleFromMaster() {
	r := gin.Default()

	r.POST("/upgrade", func(c *gin.Context) {
		n.saveUpgradeRepo(c)
	})
	r.POST("/worker", func(c *gin.Context) {
		n.watchWorkers(c)
	})

	r.Run(n.port)
}

// Watch the changes of workers' number.
// When workers increase or decrease or keep-alive, should request this API.
func (n *NodeKeeper) watchWorkers(c *gin.Context) {
	var workerInfo node.WorkerInfo
	err := c.ShouldBindBodyWith(&workerInfo, binding.JSON)
	if err != nil {
		c.String(
			http.StatusBadRequest,
			fmt.Sprintf("bad worker-info data struct: %s", err.Error()),
		)
		return
	}
	n.info.WorkersStatus[workerInfo.ID] = node.WorkerStatus{
		Info:   workerInfo,
		Online: true,
	}
	err = n.NortifyMaster()
	if err != nil {
		c.String(
			http.StatusInternalServerError,
			fmt.Sprintf("nortify master error: %s", err.Error()),
		)
		return
	}
	c.String(http.StatusOK, "")
}

func (n *NodeKeeper) saveUpgradeRepo(c *gin.Context) {
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

	c.String(http.StatusOK, "upload file succeed")
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
	req, err := http.NewRequest(http.MethodGet, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	return cli.Do(req)
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
