package error_handle

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

func LogfIfErr(err error, format string) {
	if err != nil {
		logrus.Errorf(format, err)
	}
}

func BadReqErrStr(tripodName, callName string, err error) string {
	return fmt.Sprintf("find Tripod(%s) Call(%s) error: %s", tripodName, callName, err.Error())
}
