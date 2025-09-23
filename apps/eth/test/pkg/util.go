package pkg

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sirupsen/logrus"
)

func generatePrivateKey() (string, string) {
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", ""
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)

	publicKey := privateKey.Public()
	publicKeyECDSA, _ := publicKey.(*ecdsa.PublicKey)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

	return hexutil.Encode(privateKeyBytes)[2:], address
}

func sendRequest(hostAddress string, dataString string) ([]byte, error) {
	resp, err := sendSingleRequest(hostAddress, dataString)
	if err == nil {
		return resp, nil
	}
	logrus.Infof("send request got Err:%v", err)
	for {
		time.Sleep(10 * time.Millisecond)
		resp, err = sendSingleRequest(hostAddress, dataString)
		if err == nil {
			break
		}
	}
	return resp, nil
}

func sendSingleRequest(hostAddress string, dataString string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s", hostAddress), bytes.NewBuffer([]byte(dataString)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code err: %v", resp.StatusCode)
	}
	return data, nil
}
