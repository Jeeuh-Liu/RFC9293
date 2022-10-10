package network

import (
	"time"
)

// Broadcast RIP through LinkInterface
func (node *Node) RIPRespDaemon() {
	for {
		cli := NewCLI(RIPRespBroadcast, 0, []byte{}, "")
		node.NodeCLIChan <- cli
		time.Sleep(5 * time.Second)
	}
}

func (node *Node) RIPReqDaemon() {
	cli := NewCLI(RIPReqBroadcast, 0, []byte{}, "")
	node.NodeCLIChan <- cli
}
