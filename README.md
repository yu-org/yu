# 禹

Yu is a highly customizable blockchain framework.

## Introduction
By using Yu, you can customize three levels to develop your own blockchain. The `Tripod` is for developers to 
customize their own bussiness.     
First level is define  `Exection` and `Query` on chain.  
Second level is define `blockchain lifecycle`. ( including customizable Consensus Algorithm )  
Third level is define `basic components`, such as `block data structures`, `blockchain`, `blockbase`, `txpool`. 
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

#### Examples

[Asset Tripod](https://github.com/Lawliet-Chan/yu/blob/master/apps/asset)  
`Asset Tripod` imitates an Asset function, it has `transfer accounts`, `create accounts`.  
`QueryBalance` queries someone's account balance. It implements type func `Query`.
```
func (a *Asset) QueryBalance(ctx *context.Context, env *ChainEnv, _ Hash) (interface{}, error) {
	account := ctx.GetAddress("account")
	amount := a.getBalance(env, account)
	return amount, nil
}
```  
`CreateAccount` creates an account. It implements type func `Execution`.  
`EmitEvent` will emit an event out of the chain.  
The error returned will emit out of the chain.
```
func (a *Asset) CreateAccount(ctx *context.Context, env *ChainEnv) error {
	addr := ctx.Caller
	amount := ctx.GetUint64("amount")

	if a.exsitAccount(env, addr) {
		_ = ctx.EmitEvent("Account Exists!")
		return nil
	}

	a.setBalance(env, addr, Amount(amount))
	_ = ctx.EmitEvent("Account Created Success!")
	return nil
}
```  

We need use `SetExecs` and `SetQueries` to set `Execution` and `Query` into `Asset Tripod`.
```
func NewAsset(tokenName string) *Asset {
	df := NewDefaultTripod("asset")

	a := &Asset{df, tokenName}
	a.SetExecs(a.Transfer, a.CreateAccount)
	a.SetQueries(a.QueryBalance)

	return a
}
```  
And set `Tripods` into `land`.
```
    assetTripod := asset.NewAsset("YuCoin")
    land.SetTripods(assetTripod)
```

[Pow Tripod](https://github.com/Lawliet-Chan/yu/blob/master/apps/pow/pow.go)  
`Pow Tripod` imitates a Consensus algorithm for proof of work. It customizes the lower-level code.
```

```

### Overall Structure
