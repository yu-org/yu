package error_handle

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

func LogfIfErr(err error, format string, args ...interface{}) {
	if err != nil {
		logrus.Errorf(format, err.Error(), args)
	}
}

func FindNoCallStr(tripodName, callName string, err error) string {
	return fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error())
}

func BadReqHttpResp(w http.ResponseWriter, reason string) {
	//w.WriteHeader(http.StatusBadRequest)
	//w.Write([]byte(reason))
	logrus.Panic(reason)
}

func ServerErrorHttpResp(w http.ResponseWriter, reason string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(reason))
}
