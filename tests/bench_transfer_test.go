package tests

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/apps/asset"
	"github.com/yu-org/yu/apps/poa"
	"github.com/yu-org/yu/common"
	"github.com/yu-org/yu/core/keypair"
	"github.com/yu-org/yu/core/startup"
	cliAsset "github.com/yu-org/yu/example/client/asset"
	"github.com/yu-org/yu/example/client/callchain"
	"go.uber.org/atomic"
	"net/http"
	"os"
	"sync"
	"testing"
	"time"
)

func TestTPS(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	go runChainForTPS(&wg)
	time.Sleep(2 * time.Second)
	benchmark(t)
	wg.Wait()
}

func runChainForTPS(wg *sync.WaitGroup) {

	poaCfg := poa.DefaultCfg(0)
	poaCfg.PackNum = 10000
	yuCfg := startup.InitDefaultKernelConfig()
	// yuCfg.MaxBlockNum = 10
	yuCfg.IsAdmin = true
	yuCfg.Txpool.PoolSize = 10000000

	// reset the history data
	os.RemoveAll(yuCfg.DataDir)

	assetTri := asset.NewAsset("yu-coin")
	poaTri := poa.NewPoa(poaCfg)

	chain := startup.InitDefaultKernel(yuCfg, poaTri, assetTri)
	chain.Startup()

	wg.Done()
}

type pair struct {
	pub keypair.PubKey
	prv keypair.PrivKey
}

var counter = atomic.NewInt64(0)

func benchmark(t *testing.T) {
	var users []pair
	for i := 0; i < 100; i++ {
		pub, priv := keypair.GenSrKey()
		users = append(users, pair{
			prv: priv,
			pub: pub,
		})
	}

	go caculateTPS(t)
	sub, err := callchain.NewSubscriber()
	if err != nil {
		logrus.Fatal(err)
	}
	go sub.SubEvent(nil)

	start := time.Now()

	for _, user := range users {
		cliAsset.CreateAccount(user.prv, user.pub, 1_0000_0000)
		counter.Inc()
		time.Sleep(10 * time.Microsecond)
	}

	logrus.Infof("create accounts (%d) cost %d ms", len(users), time.Since(start).Milliseconds())

	for n := 0; n < 30; n++ {
		for i, user := range users {
			var to common.Address
			if i == len(users)-1 {
				to = users[0].pub.Address()
			} else {
				to = users[i+1].pub.Address()
			}

			cliAsset.TransferBalance(user.prv, user.pub, to, 10, 0)
			counter.Inc()
			time.Sleep(10 * time.Microsecond)
		}
	}

	http.Get("http://localhost:7999/api/admin/stop")
}

func caculateTPS(t *testing.T) {
	var sec int64 = 0
	for {
		select {
		case <-time.Tick(time.Second):
			sec++
			t.Log("TPS: ", counter.Load()/sec)
		}
	}
}
