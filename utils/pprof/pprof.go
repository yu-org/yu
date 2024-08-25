package pprof

import (
	"github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof"
)

func StartPProf(address string) {
	go func() {
		err := http.ListenAndServe(address, nil)
		if err != nil {
			logrus.Error("Failure in running pprof server: ", err)
		}
	}()
}
