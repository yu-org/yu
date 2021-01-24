package node_keeper

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"yu/config"
)

const BinaryFileType = ".zip"

type NodeKeeper struct {
	binaryDir string
	port      string
}

func NewNodeKeeper(cfg *config.NodeKeeperConf) *NodeKeeper {

	return &NodeKeeper{
		binaryDir: cfg.BinaryDir,
		port:      ":" + cfg.ServesPort,
	}
}

func (n *NodeKeeper) handleFromMaster() {
	r := gin.Default()

	r.POST("/upgrade", func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("get http form error: %s", err.Error()))
			return
		}
		files := form.File["files"]
		for _, file := range files {
			fname := file.Filename
			if !strings.HasSuffix(fname, BinaryFileType) {
				c.String(
					http.StatusBadRequest,
					fmt.Sprintf("the type of file(%s) is wrong", fname),
				)
				return
			}
			err = c.SaveUploadedFile(file, n.binaryDir+"/"+fname)
			if err != nil {
				c.String(
					http.StatusInternalServerError,
					fmt.Sprintf("save file(%s) error: %s", fname, err.Error()),
				)
				return
			}
		}
		c.String(http.StatusOK, "save files succeed")
	})

	r.Run(n.port)
}
