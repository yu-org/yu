package node

type NodeType = uint

const (
	Master NodeType = iota
	Worker
)
