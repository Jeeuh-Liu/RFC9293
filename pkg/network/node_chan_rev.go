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
	case proto.CLI_LI:
		node.HandlePrintInterfaces()
		fmt.Printf("> ")
	case proto.CLI_SETUP:
		node.HandleSetUp(nodeCLI.ID)
		fmt.Printf("> ")
	case proto.CLI_SETDOWN:
		node.HandleSetDown(nodeCLI.ID)
		fmt.Printf("> ")
	case proto.CLI_QUIT:
		node.HandleQuit()
		fmt.Printf("> ")
	case proto.CLI_LR:
		node.HandlePrintRoutes()
		fmt.Printf("> ")
	case proto.MESSAGE_SENDPKT:
		node.HandleSendPacket(nodeCLI.DestIP, nodeCLI.ProtoID, nodeCLI.Msg)
		fmt.Printf("> ")
	case proto.CLI_LIFILE:
		node.HandlePrintInterfacesToFile(nodeCLI.Filename)
		fmt.Printf("> ")
	case proto.CLI_LRFILE:
		node.HandlePrintRoutesToFile(nodeCLI.Filename)
		fmt.Printf("> ")
	}
}

func (node *Node) HandleNodeBC(nodeBC *proto.NodeBC) {
	switch nodeBC.OpType {
	case proto.MESSAGE_BCRIPREQ:
		node.HandleBroadcastRIPReq()
	case proto.MESSAGE_BCRIPRESP:
		node.HandleBroadcastRIPResp()
	}
}

func (node *Node) HandleNodeEx(nodeEx *proto.NodeEx) {
	switch nodeEx.OpType {
	case proto.MESSAGE_ROUTEEX:
		node.HandleRouteEx(nodeEx.DestIP)
	}
}

func (node *Node) HandleNodePktOp(nodePktOp *proto.NodePktOp) {
	switch nodePktOp.OpType {
	case proto.MESSAGE_REVPKT:
		node.HandleReceivePacket(nodePktOp.Bytes.([]byte), nodePktOp.DestIP)
	}
}
