package ethrpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"github.com/sourcegraph/conc"
	"github.com/yu-org/yu/core/kernel"

	"github.com/reddio-com/reddio/evm"
)

const SolidityTripod = "solidity"

type EthRPC struct {
	chain     *kernel.Kernel
	cfg       *evm.GethConfig
	srv       *http.Server
	rpcServer *rpc.Server
}

func StartupEthRPC(chain *kernel.Kernel, cfg *evm.GethConfig) {
	if cfg.EnableEthRPC {
		rpcSrv, err := NewEthRPC(chain, cfg)
		if err != nil {
			logrus.Fatalf("init EthRPC server failed, %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			defer cancel()
			err = rpcSrv.Serve(ctx)
			if err != nil {
				logrus.Errorf("starknetRPC serves failed, %v", err)
			}
		}()
	}
}

func NewEthRPC(chain *kernel.Kernel, cfg *evm.GethConfig) (*EthRPC, error) {
	s := &EthRPC{
		chain:     chain,
		cfg:       cfg,
		rpcServer: rpc.NewServer(),
	}
	logrus.Debug("Start EthRpc at ", net.JoinHostPort(cfg.EthHost, cfg.EthPort))
	backend := &EthAPIBackend{
		allowUnprotectedTxs: true,
		chain:               chain,
		ethChainCfg:         cfg.ChainConfig,
	}
	backend.gasPriceCache = NewEthGasPrice(backend)

	apis := GetAPIs(backend)
	for _, api := range apis {
		err := s.rpcServer.RegisterName(api.Namespace, api.Service)
		if err != nil {
			return nil, err
		}
	}

	mux := http.NewServeMux()
	mux.Handle("/", logRequestResponse(s.rpcServer))

	s.srv = &http.Server{
		Addr:        net.JoinHostPort(cfg.EthHost, cfg.EthPort),
		Handler:     cors.Default().Handler(mux),
		ReadTimeout: 30 * time.Second,
	}

	return s, nil
}

func (s *EthRPC) Serve(ctx context.Context) error {
	errCh := make(chan error)
	defer close(errCh)

	var wg conc.WaitGroup
	defer wg.Wait()
	wg.Go(func() {
		if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	})

	select {
	case <-ctx.Done():
		return s.srv.Shutdown(context.Background())
	case err := <-errCh:
		return err
	}
}

func logRequestResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//ip := getIP(r)
		bodyBytes, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
		//logrus.Infof("[API] IP,request: %s, %s", ip, string(bodyBytes))

		rec := &responseRecorder{ResponseWriter: w, body: &bytes.Buffer{}}
		next.ServeHTTP(rec, r)
		logrus.Debugf("[API] Request:  %s", string(bodyBytes))
		//logrus.Debugf("[API] IP: %s", ip)
		logrus.Debugf("[API] Response: %s", rec.body.String())
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func getIP(r *http.Request) string {
	cfConnectingIP := r.Header.Get("CF-Connecting-IP")
	if cfConnectingIP != "" {
		return cfConnectingIP
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	xri := r.Header.Get("X-Real-Ip")
	if xri != "" {
		return xri
	}

	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}
