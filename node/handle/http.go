package handle

//
//import (
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"io"
//	"io/ioutil"
//	"net/http"
//	. "yu/common"
//	"yu/tripod"
//	"yu/txn"
//	"yu/txpool"
//)
//
//// const PARAMS_KEY = "params"
//
//func PutHttpInTxpool(c *gin.Context, txPool txpool.ItxPool, broadcastChan chan<- txn.IsignedTxn) {
//	var (
//		params JsonString
//		err    error
//	)
//	if c.Request.Method == http.MethodPost {
//		params, err = readPostBody(c.Request.Body)
//		if err != nil {
//			c.AbortWithError(http.StatusBadRequest, err)
//			return
//		}
//	} else {
//		params = c.GetString(PARAMS_KEY)
//	}
//	err = putTxpool(c.Request, params, txPool, broadcastChan)
//	if err != nil {
//		c.String(
//			http.StatusInternalServerError,
//			fmt.Sprintf("Put Execution into TxPool error: %s", err.Error()),
//		)
//		return
//	}
//	c.String(http.StatusOK, "")
//}
//
//func DoHttpQryCall(c *gin.Context, land *tripod.Land) {
//	var (
//		params JsonString
//		err    error
//	)
//	if c.Request.Method == http.MethodPost {
//		params, err = readPostBody(c.Request.Body)
//		if err != nil {
//			c.AbortWithError(http.StatusBadRequest, err)
//			return
//		}
//	} else {
//		params = c.GetString(PARAMS_KEY)
//	}
//	err = doQryCall(c.Request, params, land)
//	if err != nil {
//		c.String(
//			http.StatusInternalServerError,
//			fmt.Sprintf("Query error: %s", err.Error()),
//		)
//		return
//	}
//	c.String(http.StatusOK, "")
//}
//
//func readPostBody(body io.ReadCloser) (JsonString, error) {
//	byt, err := ioutil.ReadAll(body)
//	return JsonString(byt), err
//}
