package node

import (
	"bufio"
	"github.com/libp2p/go-libp2p-core/network"
)

func (m *MasterNode) handleStream(s network.Stream) {
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))
	go ReadFromNetwork(rw)
	go WriteToNetwork(rw)
}

func ReadFromNetwork(rw *bufio.ReadWriter) {

}

func WriteToNetwork(rw *bufio.ReadWriter) {

}
