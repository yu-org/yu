# 禹

Yu is a highly customizable blockchain framework.

## Introduction
By using Yu, you can customize two levels to develop your own blockchain.  
One level is define  `Exection` and `Query` on chain.  
Two level is define `blockchain lifecycle`. ( including customizable Consensus Algorithm )
- Define your `Exection` and `Query` on  chain.  
`Execution` is like `Transaction` in Ethereum but not only for transfer of Token, it changes the state on the chain and must be consensus on all nodes.  
`Query` is like `query` in Ethereum, it doesn't change state, just query some data from the chain.  

```
type (
	Execution func(*context.Context, *chain_env.ChainEnv) error
	
	Query func(*context.Context, *chain_env.ChainEnv, common.Hash) (respObj interface{}, err error)
)
```
- Define Your `blockchain lifecycle`, this function is in `Tripod` interface.  
`InitChain` defines bussiness when the blockchain starts up. You should use it to define `Genesis Block`.  
`StartBlock` defines bussiness when a new block starts. In this func, you can set some attributes( including package txns from txpool, mining ) in the block,
then you should tell the framework whether broadcast the block to other nodes or not.    
`EndBlock` defines bussiness when all nodes accept the new block, usually we execute the txns of new block and append  block into the chain.  
`FinalizeBlock` defines bussiness when the block is finalized in the chain by all nodes.
 
```
type Tripod interface {

	......

	InitChain(env *ChainEnv, land *Land) error

	StartBlock(block IBlock, env *ChainEnv, land *Land) (needBroadcast bool, err error)

	EndBlock(block IBlock, env *ChainEnv, land *Land) error

	FinalizeBlock(block IBlock, env *ChainEnv, land *Land) error
}
```

#### Quick Start





### Overall Structure
