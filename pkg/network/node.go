package network

import (
	"fmt"
	"tcpip/pkg/proto"
	"tcpip/pkg/tcp"
)

// The driver program

type Node struct {
	// Channel
	NodeCLIChan   chan *proto.NodeCLI   // Receive CLI from user
	NodeBCChan    chan *proto.NodeBC    // Broadcast RIP
	NodeExChan    chan *proto.NodeEx    // Handle expiration of route
	NodePktOpChan chan *proto.NodePktOp // Receive msg from link interface
	RT            *RoutingTable
	socketTable   *tcp.SocketTable
	segRecvChan   chan *proto.Segment
	segSendChan   chan *proto.SegmentMsg
}

func (node *Node) Make(args []string) {
	// Initialize Channel
	node.NodeCLIChan = make(chan *proto.NodeCLI)
	node.NodeBCChan = make(chan *proto.NodeBC)
	node.NodeExChan = make(chan *proto.NodeEx)
	node.NodePktOpChan = make(chan *proto.NodePktOp)

	node.RT = &RoutingTable{}
	node.RT.Make(args, node.NodePktOpChan, node.NodeExChan)

	node.socketTable = tcp.NewSocketTable()
	node.segRecvChan = make(chan *proto.Segment)
	node.segSendChan = make(chan *proto.SegmentMsg)

	// Receive CLI
	go node.ScanClI()
	// Broadcast RIP Request once
	go node.RIPReqDaemon()
	// Broadcast RIP Resp periodically
	go node.RIPRespDaemon()

	go node.handleTCP()
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
