package node

import (
	"context"
	"github.com/libp2p/go-libp2p"
)

type Node struct {
	typ NodeType

}

func NewNode() *Node {
	ctx, cancel := context.WithCancel(context.Background())
	libp2p.New(ctx)
	libp2p.Identity()
}