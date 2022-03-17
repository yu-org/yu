# 禹

Yu is a highly customizable blockchain framework.  

[Book](https://yu-org.github.io/yu-docs/en/)  
[中文文档](https://yu-org.github.io/yu-docs/zh/)

### Overall Structure
![image](yu_flow_chart.png)

## Introduction
By using Yu, you can customize three levels to develop your own blockchain. The `Tripod` is for developers to 
customize their own business.     
First level is define  `Execution` and `Query` on chain.  
Second level is define `blockchain lifecycle`. ( including customizable Consensus Algorithm )  
Third level is define `basic components`, such as `block data structures`, `blockchain`, `yudb`, `txpool`. 
- Define your `Execution` and `Query` on  chain.  
`Execution` is like `Transaction` in Ethereum but not only for transfer of Token, it changes the state on the chain and must be consensus on all nodes.  
`Query` is like `query` in Ethereum, it doesn't change state, just query some data from the chain.  
`P2pHandler` is a p2p server handler. You can define the services in P2P server. Just like TCP handler.  

```go
type (
    Execution func(ctx *context.Context, currentBlock *types.CompactBlock) error
	
    Query func(ctx *context.Context, blockHash Hash) (respObj interface{}, err error)

    P2pHandler func([]byte) ([]byte, error)
)
```
- Define Your `blockchain lifecycle`, this function is in `Tripod` interface.  
`CheckTxn` defines the rules for checking transactions(Executions) before inserting txpool.  
`VerifyBlock` defines the rules for verifying blocks.   
`InitChain` defines business when the blockchain starts up. You should use it to define `Genesis Block`.  
`StartBlock` defines business when a new block starts. In this func, you can set some attributes( including pack txns from txpool, mining ) in the block,
then you should tell the framework whether broadcast the block to other nodes or not.    
`EndBlock` defines business when all nodes accept the new block, usually we execute the txns of new block and append  block into the chain.  
`FinalizeBlock` defines business when the block is finalized in the chain by all nodes.
 
```go
type Tripod interface {

    ......
    
    CheckTxn(*txn.SignedTxn)    

    VerifyBlock(block *types.CompactBlock) bool

    InitChain() error

    StartBlock(block *types.CompactBlock) error

    EndBlock(block *types.CompactBlock) error

    FinalizeBlock(block *types.CompactBlock) error
}
```

#### Examples

[Asset Tripod](https://github.com/yu-org/yu/blob/master/apps/asset)  
`Asset Tripod` imitates an Asset function, it has `transfer accounts`, `create accounts`.  
`QueryBalance` queries someone's account balance. It implements type func `Query`.
```go
func (a *Asset) QueryBalance(ctx *context.Context, _ Hash) (interface{}, error) {
    account := ctx.GetAddress("account")
    if !a.existAccount(account) {
        return nil, AccountNotFound(account)
    }
    amount := a.getBalance(account)
    return amount, nil
}
```  
`CreateAccount` creates an account. It implements type func `Execution`.  
`EmitEvent` will emit an event out of the chain.  
The error returned will emit out of the chain.
```go
func (a *Asset) CreateAccount(ctx *context.Context, _ *CompactBlock) error {
    addr := ctx.Caller
	amount := big.NewInt(int64(ctx.GetUint64("amount")))

    if a.existAccount(addr) {
    _ = ctx.EmitEvent("Account Exists!")
    return nil
    }

    a.setBalance(addr, amount)
    _ = ctx.EmitEvent("Account Created Success!")
    return nil
}
```  

We need use `SetExec` and `SetQueries` to set `Execution` and `Query` into `Asset Tripod`.  
When we set a `Execution`, we need declare how much `Lei`(耜) it consumes. (`Lei` is the same as `gas` in `ethereum` )
```go
func NewAsset(tokenName string) *Asset {
    df := NewDefaultTripod("asset")

    a := &Asset{df, tokenName}
    a.SetExec(a.Transfer, 100).SetExec(a.CreateAccount, 10)
    a.SetQueries(a.QueryBalance)

    return a
}
```  
Finally set `Asset Tripod` into `land` in `main func`. 
```go
func main() {
    startup.StartUp(pow.NewPow(1024), asset.NewAsset("YuCoin"))
}
```

[Poa Tripod](https://github.com/yu-org/yu/blob/master/apps/poa/poa.go)  
`Pow Tripod` imitates a Consensus algorithm for proof of authority. It customizes the lower-level code.
- Start a new block  
If there are no verified blocks from P2P network, we pack some txns, mine a new block and broadcast it to P2P network.
```go
func (h *Poa) StartBlock(block *CompactBlock) error {
    ......
	
    // Get a leader who produce the block of this round. 
    miner := h.CompeteLeader(block.Height)
    logrus.Debugf("compete a leader(%s) in round(%d)", miner.String(), block.Height)
	
    // If it is not local node for this round, use other node's block and skip follows.
    if miner != h.LocalAddress() {
        if h.useP2pOrSkip(block) {
            logrus.Infof("--------USE P2P Height(%d) block(%s) miner(%s)",
            block.Height, block.Hash.String(), ToHex(block.MinerPubkey))
            return nil
        }
    }
	
    // Pack transactions(Executions) from Txpool. 
    txns, err := h.env.Pool.Pack(3000)
    if err != nil {
        return err
    }
    // Make blockHash from trasactions. 
    hashes := FromArray(txns...).Hashes()
    block.TxnsHashes = hashes
    txnRoot, err := MakeTxnRoot(txns)
    if err != nil {
        return err
    }
    block.TxnRoot = txnRoot
	
    ......

    // signs block
    block.MinerSignature, err = h.myPrivKey.SignData(block.Hash.Bytes())
    if err != nil {
        return err
    }
    block.MinerPubkey = h.myPubkey.BytesWithType()
	
    // Reset Txpool for the next block.
    err = h.env.Pool.Reset(block)
    if err != nil {
        return err
    }
    
    ......
	
    // Publish the block to P2P so that other nodes get it.
    return h.env.P2pNetwork.PubP2P(StartBlockTopic, rawBlockByt)
}
```
- End the block  
We execute the txns of the block and append the block into the chain.
```go
func (h *Pow) EndBlock(block *CompactBlock) error {
    ......
    // Execute all transactions(executions) of this block.
    err := h.env.Execute(block)
    if err != nil {
        return err
    }

    // Append the block into the chain.
    err = chain.AppendBlock(block)
    if err != nil {
        return err
    }  
    ......
}
```

- Finalize the block   
```go
func (h *Poa) FinalizeBlock(block *CompactBlock) error {
    return h.env.Chain.Finalize(block.Hash)
}
```


Same as `Asset Tripod` , finally set `Pow Tripod` into `land` in `main function`.    
```go
func main() {
    startup.StartUp(poa.NewPoa(myPubkey, myPrivkey, validatorsAddrs), asset.NewAsset("YuCoin"))
}
```

