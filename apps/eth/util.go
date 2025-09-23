package eth

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/holiman/uint256"
	yu_common "github.com/yu-org/yu/common"
)

func ConvertHashToYuHash(hash common.Hash) (yu_common.Hash, error) {
	var yuHash [yu_common.HashLen]byte
	if len(hash.Bytes()) == yu_common.HashLen {
		copy(yuHash[:], hash.Bytes())
		return yuHash, nil
	} else {
		return yu_common.Hash{}, errors.New(fmt.Sprintf("Expected hash to be 32 bytes long, but got %d bytes", len(hash.Bytes())))
	}
}

func ConvertBigIntToUint256(b *big.Int) *uint256.Int {
	if b == nil {
		return nil
	}
	u, _ := uint256.FromBig(b)
	return u
}

func ObjToJson(obj interface{}) string {
	byt, err := json.Marshal(obj)
	if err != nil {
		return ""
	}
	return string(byt)
}

func ValidateTxHash(hash string) bool {
	if len(hash) != 66 || !strings.HasPrefix(hash, "0x") {
		return false
	}

	if isAllZero(hash) {
		return false
	}

	if countLeadingZeros(hash) > 16 {
		return false
	}

	return true
}

func isAllZero(hash string) bool {
	for _, c := range hash[2:] {
		if c != '0' {
			return false
		}
	}
	return true
}

func countLeadingZeros(hash string) int {
	count := 0
	for _, c := range hash[2:] {
		if c != '0' {
			break
		}
		count++
	}
	return count
}
