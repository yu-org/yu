package interfaces

import (
	. "yu/blockchain"
)

type Pod interface {

	OnInitialize(blockNum BlockNum) error

	OnFinalize(blockNum BlockNum) error
}
