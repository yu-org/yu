package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/example/client/asset"
	. "github.com/yu-org/yu/example/client/callchain"
	"go.uber.org/atomic"
	"time"
)

type pair struct {
	pub PubKey
	prv PrivKey
}

var counter = atomic.NewInt64(0)

func main() {
	var users []pair
	for i := 0; i < 100; i++ {
		pub, priv := GenSrKey()
		users = append(users, pair{
			prv: priv,
			pub: pub,
		})
	}

	go caculateTPS()
	go SubEvent()

	start := time.Now()

	for _, user := range users {
		CreateAccount(Websocket, user.prv, user.pub, 1_0000_0000)
		counter.Inc()
		time.Sleep(10 * time.Microsecond)
	}

	logrus.Infof("create accounts (%d) cost %d ms", len(users), time.Since(start).Milliseconds())

	for j := 0; j < 100; j++ {
		for i, user := range users {
			var to common.Address
			if i == len(users)-1 {
				to = users[0].pub.Address()
			} else {
				to = users[i+1].pub.Address()
			}

			TransferBalance(Websocket, user.prv, user.pub, to, 10, 0)
			counter.Inc()
			time.Sleep(100 * time.Microsecond)
		}
		logrus.Info("----- transfer one turn")
	}
	select {}
}

func caculateTPS() {
	var sec int64 = 0
	for {
		select {
		case <-time.Tick(time.Second):
			sec++
			logrus.Info("TPS: ", counter.Load()/sec)
		}
	}
}
