package node_keeper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"yu/config"
	"yu/utils/compress"
)

const BinaryFileType = ".zip"

type NodeKeeper struct {
	repos []*Repo
	dir   string
	port  string
}

func NewNodeKeeper(cfg *config.NodeKeeperConf) *NodeKeeper {

	return &NodeKeeper{
		dir:  cfg.Dir,
		port: ":" + cfg.ServesPort,
	}
}

func (n *NodeKeeper) handleFromMaster() {
	r := gin.Default()

	r.POST("/upgrade", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("upload file error: %s", err.Error()))
			return
		}

		fname := file.Filename
		if !strings.HasSuffix(fname, BinaryFileType) {
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
		err = n.convertToRepo(zipFileName)
		if err != nil {
			c.String(
				http.StatusInternalServerError,
				fmt.Sprintf("convert file(%s) to repo error: %s", fname, err.Error()),
			)
		}

		c.String(http.StatusOK, "upload file succeed")
	})

	r.Run(n.port)
}

func (n *NodeKeeper) convertToRepo(zipFileName string) error {
	files, err := compress.UnzipFile(zipFileName, n.dir)
	if err != nil {
		return err
	}
	repoDir := strings.TrimSuffix(zipFileName, BinaryFileType)
	err = os.MkdirAll(repoDir, os.ModePerm)
	if err != nil {
		return err
	}

	n.repos = append(n.repos, NewRepo(files))
	return os.Remove(zipFileName)
}

func ReposFromDir(dir string) []*Repo {

}
