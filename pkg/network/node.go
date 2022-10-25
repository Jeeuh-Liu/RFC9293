package network

import (
	"fmt"
	"tcpip/pkg/proto"
)

// The driver program

type Node struct {
	// Channel
	NodeCLIChan   chan *proto.NodeCLI   // Receive CLI from user
	NodeBCChan    chan *proto.NodeBC    // Broadcast RIP
	NodeExChan    chan *proto.NodeEx    // Handle expiration of route
	NodePktOpChan chan *proto.NodePktOp // Receive msg from link interface
	RT            *RoutingTable
}

func (node *Node) Make(args []string) {
	// Initialize Channel
	node.NodeCLIChan = make(chan *proto.NodeCLI)
	node.NodeBCChan = make(chan *proto.NodeBC)
	node.NodeExChan = make(chan *proto.NodeEx)
	node.NodePktOpChan = make(chan *proto.NodePktOp)

	node.RT = &RoutingTable{}
	node.RT.Make(args, node.NodePktOpChan, node.NodeExChan)

	// Receive CLI
	go node.ScanClI()
	// Broadcast RIP Request once
	go node.RIPReqDaemon()
	// Broadcast RIP Resp periodically
	go node.RIPRespDaemon()
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
