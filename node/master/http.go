package master

import (
	"github.com/gin-gonic/gin"
	"io"
	"io/ioutil"
	"net/http"
	. "yu/common"
	"yu/context"
	. "yu/node"
	. "yu/txn"
	. "yu/utils/error_handle"
)

const PARAMS_KEY = "params"

func (m *Master) HandleHttp() {
	r := gin.Default()

	if m.RunMode == MasterWorker {
		r.POST(RegisterNodeKeepersPath, func(c *gin.Context) {
			m.registerNodeKeepers(c)
		})
	}

	// GET request
	r.GET(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.GET(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	// POST request
	r.POST(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.POST(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	// PUT request
	r.PUT(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.PUT(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	// DELETE request
	r.DELETE(ExecApiPath, func(c *gin.Context) {
		m.handleHttpExec(c)
	})
	r.DELETE(QryApiPath, func(c *gin.Context) {
		m.handleHttpQry(c)
	})

	r.Run(m.httpPort)
}

func (m *Master) handleHttpExec(c *gin.Context) {
	params, err := getHttpJsonParams(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	tripodName, callName, stxn, err := getExecInfoFromReq(c.Request, params)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	switch m.RunMode {
	case MasterWorker:
		ip, name, err := m.findWorkerIpAndName(tripodName, callName, ExecCall)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				FindNoCallStr(tripodName, callName, err),
			)
			return
		}

		fmap := make(map[string]*TxnsAndWorkerName)
		fmap[ip] = &TxnsAndWorkerName{
			Txns:       FromArray(stxn),
			WorkerName: name,
		}
		err = m.forwardTxnsForCheck(fmap)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		err = m.txPool.Insert(name, stxn)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	case LocalNode:
		err = m.txPool.Insert("", stxn)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
	}

	err = m.pubUnpackedTxns(FromArray(stxn))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
	}
}

func (m *Master) handleHttpQry(c *gin.Context) {
	params, err := getHttpJsonParams(c)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	qcall, err := getQryInfoFromReq(c.Request, params)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	switch m.RunMode {
	case MasterWorker:
		var ip string
		ip, err = m.findWorkerIP(qcall.TripodName, qcall.QueryName, QryCall)
		if err != nil {
			c.String(
				http.StatusBadRequest,
				FindNoCallStr(qcall.TripodName, qcall.QueryName, err),
			)
			return
		}
		forwardQueryToWorker(ip, c.Writer, c.Request)
	case LocalNode:
		pubkey, err := GetPubkey(c.Request)
		if err != nil {

			return
		}
		ctx, err := context.NewContext(pubkey.Address(), qcall.Params)
		if err != nil {
			c.String(http.StatusBadRequest, err.Error())
			return
		}

		endBlock, err := m.chain.GetEndBlock()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		respObj, err := m.land.Query(qcall, ctx, m.GetEnv(endBlock))
		if err != nil {
			c.String(
				http.StatusBadRequest,
				FindNoCallStr(qcall.TripodName, qcall.QueryName, err),
			)
			return
		}
		c.JSON(http.StatusOK, respObj)
	}

}

func readPostBody(body io.ReadCloser) (JsonString, error) {
	byt, err := ioutil.ReadAll(body)
	return JsonString(byt), err
}
