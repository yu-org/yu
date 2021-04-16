package master

import . "yu/common"

type HandShake struct {
	GenesisBlockHash Hash

	// When POW, these two will be always 0 and null.
	FinalizeHeight    BlockNum
	FinalizeBlockHash Hash

	EndHeight    BlockNum
	EndBlockHash Hash
}

type FetchBlocksRequest struct {
	StartHeight BlockNum
	Count       uint64
}
