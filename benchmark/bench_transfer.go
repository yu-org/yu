package main

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	. "github.com/yu-org/yu/core/keypair"
	. "github.com/yu-org/yu/example/client/asset"
	. "github.com/yu-org/yu/example/client/callchain"
	"go.uber.org/atomic"
	"sync"
	"time"
)

type pair struct {
	pub PubKey
	prv PrivKey
}

var counter = atomic.NewInt64(0)

func main() {
	var users []pair
	for i := 0; i < 1000; i++ {
		pub, priv := GenSrKey()
		users = append(users, pair{
			prv: priv,
			pub: pub,
		})
	}

	go caculateTPS()
	go SubEvent()

	wg := &sync.WaitGroup{}

	for _, user := range users {
		wg.Add(1)
		go func(user pair, wg *sync.WaitGroup) {
			CreateAccount(Websocket, user.prv, user.pub, 1_0000_0000)
			counter.Inc()
			wg.Done()
		}(user, wg)

		time.Sleep(time.Microsecond * 200)
	}
	wg.Wait()

	for {
		for i, user := range users {
			var to common.Address
			if i == len(users)-1 {
				to = users[0].pub.Address()
			} else {
				to = users[i+1].pub.Address()
			}

			go func(user pair, to common.Address) {
				TransferBalance(Websocket, user.prv, user.pub, to, 10, 1)
				counter.Inc()
			}(user, to)

			time.Sleep(time.Microsecond * 200)
		}
	}
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
