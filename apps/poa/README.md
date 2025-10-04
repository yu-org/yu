# PoA Tripod

`Poa Tripod` implements a Proof of Authority consensus algorithm for the Yu blockchain framework. It customizes the lower-level blockchain lifecycle functions to provide a robust and efficient consensus mechanism.

## Overview

Proof of Authority (PoA) is a consensus algorithm that relies on a limited number of block validators (authorities) to create new blocks and secure the network. This makes it ideal for private or consortium blockchains where trust between participants is established.

## Key Features

- **Authority-based consensus**: Only designated validators can produce blocks
- **Leader selection**: Rotating leader mechanism for block production
- **P2P synchronization**: Automatic block synchronization across the network
- **Transaction packing**: Efficient transaction collection and inclusion
- **Block finalization**: Secure block finalization process

## Implementation Details

### Start Block Process

The `StartBlock` function handles the beginning of a new block creation cycle:

```go
func (h *Poa) StartBlock(block *Block) {
    // Get a leader who produce the block of this round. 
    miner := h.CompeteLeader(block.Height)
    logrus.Debugf("compete a leader(%s) in round(%d)", miner.String(), block.Height)
    
    // If it is not local node for this round, use other node's block and skip follows.
    if miner != h.LocalAddress() {
        if h.useP2pOrSkip(block) {
            logrus.Infof("--------USE P2P Height(%d) block(%s) miner(%s)",
            block.Height, block.Hash.String(), ToHex(block.MinerPubkey))
            return
        }
    }
    
    // Pack transactions(Writings) from Txpool. 
    txns, err := h.env.Pool.Pack(3000)
    if err != nil {
        logrus.Error(err)
        return 
    }
    
    // Make blockHash from trasactions. 
    hashes := FromArray(txns...).Hashes()
    block.TxnsHashes = hashes
    txnRoot, err := MakeTxnRoot(txns)
    if err != nil {
        logrus.Error(err)
        return 
    }
    block.TxnRoot = txnRoot
    
    // signs block
    block.MinerSignature, err = h.myPrivKey.SignData(block.Hash.Bytes())
    if err != nil {
        logrus.Error(err)
        return 
    }
    block.MinerPubkey = h.myPubkey.BytesWithType()
    
    // Reset Txpool for the next block.
    err = h.env.Pool.Reset(block)
    if err != nil {
        logrus.Error(err)
        return
    }
    
    // Publish the block to P2P so that other nodes get it.
    return h.env.P2pNetwork.PubP2P(StartBlockTopic, rawBlockByt)
}
```

**Key Steps:**
1. **Leader Competition**: Determine which validator should produce the current block
2. **P2P Synchronization**: If not the current leader, synchronize with other nodes
3. **Transaction Packing**: Collect up to 3000 transactions from the transaction pool
4. **Block Construction**: Create transaction root and prepare block structure
5. **Block Signing**: Sign the block with the validator's private key
6. **Pool Reset**: Clear processed transactions from the pool
7. **P2P Broadcasting**: Broadcast the new block to the network

### End Block Process

The `EndBlock` function handles block execution and chain appending:

```go
func (h *Poa) EndBlock(block *Block) {
    // Execute all transactions(Writings) of this block.
    err := h.env.Execute(block)
    if err != nil {
        logrus.Error(err)
        return 
    }

    // Append the block into the chain.
    err = chain.AppendBlock(block)
    if err != nil {
        logrus.Error(err)
        return 
    }  
}
```

**Key Steps:**
1. **Transaction Execution**: Execute all transactions in the block
2. **Chain Append**: Add the block to the blockchain

### Finalize Block Process

The `FinalizeBlock` function handles block finalization:

```go
func (h *Poa) FinalizeBlock(block *Block) {
    h.env.Chain.Finalize(block.Hash)
}
```

**Purpose:**
- Finalizes the block in the chain
- Marks the block as immutable and permanently committed

## Usage Example

To use the PoA Tripod in your Yu application:

```go
func main() {
    startup.InitConfigFromPath("yu_conf/kernel.toml")
    startup.DefaultStartup(
        poa.NewPoa(poaConf),
        asset.NewAsset("YuCoin"),
    )
}
```

## Configuration

The PoA Tripod requires specific configuration parameters:

- **Validators**: List of authorized block producers
- **Block Interval**: Time between block production
- **Transaction Limit**: Maximum transactions per block (default: 3000)
- **P2P Settings**: Network synchronization parameters

## Benefits

- **Performance**: Fast block production with minimal computational overhead
- **Consistency**: Reliable block production through designated validators
- **Efficiency**: Low energy consumption compared to Proof of Work
- **Control**: Suitable for private and consortium networks
- **Scalability**: Handles high transaction throughput

## Use Cases

- **Private Blockchains**: Enterprise and institutional use cases
- **Consortium Networks**: Multi-organization blockchain systems
- **Development Networks**: Testing and development environments
- **Hybrid Systems**: Combining PoA with other consensus mechanisms

## Security Considerations

- **Validator Trust**: Security relies on the trustworthiness of validators
- **Key Management**: Proper private key protection is essential
- **Network Security**: P2P communication should be secured
- **Validator Rotation**: Consider implementing validator rotation mechanisms

---

For more information about the Yu framework and other Tripods, visit the [main documentation](https://yu-org.github.io/yu-docs/).
