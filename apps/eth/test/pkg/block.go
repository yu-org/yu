package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common/yerror"
	"github.com/yu-org/yu/core/types"
)

func GetDefaultBlockManager() *BlockManager {
	return &BlockManager{
		hostUrl: "localhost:7999",
	}
}

type BlockManager struct {
	hostUrl string
}

func (bm *BlockManager) StopBlockChain() {
	_, err := http.Get(fmt.Sprintf("http://%s/api/admin/stop", bm.hostUrl))
	if err != nil {
		logrus.Error("Failed to stop blockchain", err)
	}
}

func (bm *BlockManager) GetBlockTxnCountByIndex(index int) (bool, int, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/receipts_count?block_number=%v", bm.hostUrl, index))
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0, err
	}
	r := &txCountResp{}
	err = json.Unmarshal(d, &r)
	if err != nil {
		return false, 0, err
	}
	if r.ErrMsg == yerror.ErrBlockNotFound.Error() {
		return false, 0, nil
	}
	return true, r.Data, nil
}

func (bm *BlockManager) GetCurrentBlock() (*types.Block, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/block", bm.hostUrl))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := &blockResp{}
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}
	if r.ErrMsg == yerror.ErrBlockNotFound.Error() {
		return nil, err
	}
	return r.Data, nil
}

func (bm *BlockManager) GetBlockByIndex(id uint64) (*types.Block, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/api/block?number=%s", bm.hostUrl, hexutil.EncodeUint64(id)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r := &blockResp{}
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}
	if r.ErrMsg == yerror.ErrBlockNotFound.Error() {
		return nil, err
	}
	return r.Data, nil
}

type txCountResp struct {
	Code   int    `json:"code"`
	ErrMsg string `json:"err_msg"`
	Data   int    `json:"data"`
}

type blockResp struct {
	Code   int          `json:"code"`
	ErrMsg string       `json:"err_msg"`
	Data   *types.Block `json:"data"`
}
