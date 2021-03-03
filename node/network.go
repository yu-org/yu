package node

import (
	"bytes"
	"github.com/sirupsen/logrus"
	"net/http"
)

func SendHeartbeats(addrs []string, handleDead func(addr string) error) {
	for _, addr := range addrs {
		_, err := http.Get(addr + HeartbeatPath)
		if err != nil {
			logrus.Errorf("send heartbeat to (%s) error: %s", addr, err.Error())
			err = handleDead(addr)
			if err != nil {
				logrus.Errorf("handle dead node (%s) error: %s", addr, err.Error())
			}
		} else {
			logrus.Debugf("send heartbeat to (%s) succeed!", addr)
		}
	}

}

func PostRequest(url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	cli := &http.Client{}
	return cli.Do(req)
}
