package tests

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/example/client/asset"
	"github.com/yu-org/yu/example/client/callchain"
	"go.uber.org/atomic"
	"net/http"
	"sync"
	"testing"
	"time"
)

func BenchmarkTransfer(b *testing.B) {
	var wg sync.WaitGroup
	wg.Add(1)

	go runChain(&wg)
	time.Sleep(2 * time.Second)
	benchmark(b)
	wg.Wait()
}

type pair struct {
	pub keypair.PubKey
	prv keypair.PrivKey
}

var counter = atomic.NewInt64(0)

func benchmark(b *testing.B) {
	var users []pair
	for i := 0; i < 100; i++ {
		pub, priv := keypair.GenSrKey()
		users = append(users, pair{
			prv: priv,
			pub: pub,
		})
	}

	go caculateTPS(b)
	sub, err := callchain.NewSubscriber()
	if err != nil {
		logrus.Fatal(err)
	}
	go sub.SubEvent(nil)

	start := time.Now()

	for _, user := range users {
		asset.CreateAccount(user.prv, user.pub, 1_0000_0000)
		counter.Inc()
		time.Sleep(10 * time.Microsecond)
	}

	logrus.Infof("create accounts (%d) cost %d ms", len(users), time.Since(start).Milliseconds())

	for n := 0; n < b.N; n++ {
		for i, user := range users {
			var to common.Address
			if i == len(users)-1 {
				to = users[0].pub.Address()
			} else {
				to = users[i+1].pub.Address()
			}

			asset.TransferBalance(user.prv, user.pub, to, 10, 0)
			counter.Inc()
			time.Sleep(10 * time.Microsecond)
		}
	}

	http.Get("http://localhost:7999/api/admin/stop")
}

func caculateTPS(b *testing.B) {
	var sec int64 = 0
	for {
		select {
		case <-time.Tick(time.Second):
			sec++
			b.Log("TPS: ", counter.Load()/sec)
		}
	}
}
