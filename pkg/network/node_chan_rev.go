package network

import (
	"fmt"
	"tcpip/pkg/proto"
)

// Output the data of CLI
func (node *Node) ReceiveOpFromChan() {
	fmt.Printf("> ")
	for {
		select {
		case nodeCLI := <-node.NodeCLIChan:
			// fmt.Println(nodeCLI)
			node.HandleNodeCLI(nodeCLI)
		case nodeBC := <-node.NodeBCChan:
			// fmt.Println(nodeBC)
			node.HandleNodeBC(nodeBC)
		case nodeEx := <-node.NodeExChan:
			// fmt.Println(nodeEx)
			node.HandleNodeEx(nodeEx)
		case nodePktOp := <-node.NodePktOpChan:
			// fmt.Println(nodePktOp)
			node.HandleNodePktOp(nodePktOp)
		}
	}
}

func (node *Node) HandleNodeCLI(nodeCLI *proto.NodeCLI) {
	switch nodeCLI.CLIType {
	// CLI
	case proto.LI:
		node.HandlePrintInterfaces()
		fmt.Printf("> ")
	case proto.SetUpT:
		node.HandleSetUp(nodeCLI.ID)
		fmt.Printf("> ")
	case proto.SetDownT:
		node.HandleSetDown(nodeCLI.ID)
		fmt.Printf("> ")
	case proto.Quit:
		node.HandleQuit()
		fmt.Printf("> ")
	case proto.LR:
		node.HandlePrintRoutes()
		fmt.Printf("> ")
	case proto.TypeSendPacket:
		node.HandleSendPacket(nodeCLI.DestIP, nodeCLI.ProtoID, nodeCLI.Msg)
		fmt.Printf("> ")
	}
}

func (node *Node) HandleNodeBC(nodeBC *proto.NodeBC) {
	switch nodeBC.OpType {
	case proto.TypeBroadcastRIPReq:
		node.HandleBroadcastRIPReq()
	case proto.TypeBroadcastRIPResp:
		node.HandleBroadcastRIPResp()
	}
}

func (node *Node) HandleNodeEx(nodeEx *proto.NodeEx) {
	switch nodeEx.OpType {
	case proto.TypeRouteEx:
		node.HandleRouteEx(nodeEx.DestIP)
	}
}

func (node *Node) HandleNodePktOp(nodePktOp *proto.NodePktOp) {
	switch nodePktOp.OpType {
	case proto.TypeReceivePacket:
		node.HandleReceivePacket(nodePktOp.Bytes.([]byte), nodePktOp.DestIP)
	}
}
