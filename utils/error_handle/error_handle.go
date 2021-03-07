package error_handle

import "github.com/sirupsen/logrus"

func LogfIfErr(err error, format string) {
	if err != nil {
		logrus.Errorf(format, err)
	}
}
