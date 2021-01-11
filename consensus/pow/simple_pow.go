package pow


import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"math"
	"math/big"
	. "yu/blockchain"
	"yu/common"
)

const targetBits = 16

type SimplePoW struct {
	block  IBlock
	target *big.Int
}

func NewSimplePoW(b IBlock) *SimplePoW {
	target := big.NewInt(1)
	target.Lsh(target, 256-targetBits)
	return &SimplePoW{
		b,
		target,
	}
}

func (sp *SimplePoW) prepareData(nonce int) ([]byte, error) {
	hex1, err := IntToHex(sp.block.Timestamp())
	if err != nil {
		return nil, err
	}
	hex2, err := IntToHex(int64(targetBits))
	if err != nil {
		return nil, err
	}
	hex3, err := IntToHex(int64(nonce))
	if err != nil {
		return nil, err
	}
	data := bytes.Join(
		[][]byte{
			sp.block.PrevHash().Bytes(),
			sp.block.Hash().Bytes(),
			hex1,
			hex2,
			hex3,
		},
		[]byte{},
	)

	return data, nil
}

func (sp *SimplePoW) Run() (nonce int, hash common.Hash, err error) {
	var hashInt big.Int
	nonce = 0

	logrus.Info("Mining a new Block")
	for nonce < math.MaxInt64 {
		var data []byte
		data, err = sp.prepareData(nonce)
		if err != nil {
			return
		}
		hash = sha256.Sum256(data)
		if math.Remainder(float64(nonce), 100000) == 0 {
			logrus.Infof("Hash is \r%x", hash.Bytes())
		}
		hashInt.SetBytes(hash.Bytes())

		if hashInt.Cmp(sp.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	return
}

func (sp *SimplePoW) Validate() bool {
	var hashInt big.Int

	var nonce int = sp.block.Extra().(int)
	data, err := sp.prepareData(nonce)
	if err != nil {
		return false
	}
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(sp.target) == -1
}

func IntToHex(num int64) ([]byte, error) {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}
