package network

import (
	"fmt"
	"log"
	"os"
	"tcpip/pkg/proto"

	"golang.org/x/net/ipv4"
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

// ***********************************************************************************
// Handle CLI
func (node *Node) HandlePrintInterfaces() {
	node.RT.PrintInterfaces()
}

func (node *Node) HandlePrintInterfacesToFile(filename string) {
	node.RT.PrintInterfacesToFile(filename)
}

func (node *Node) HandleSetUp(id uint8) {
	node.RT.SetUp(id)
}

func (node *Node) HandleSetDown(id uint8) {
	node.RT.SetDown(id)
}

func (node *Node) HandleQuit() {
	os.Exit(0)
}

func (node *Node) HandlePrintRoutes() {
	node.RT.PrintRoutes()
}

func (node *Node) HandlePrintRoutesToFile(filename string) {
	node.RT.PrintRoutesToFile(filename)
}

func (node *Node) HandleSendPacket(destIP string, protoID int, msg string) {
	node.RT.SendPacket(destIP, protoID, msg)
}

// ***********************************************************************************
// Handle BroadcastRIP
func (node *Node) HandleBroadcastRIPReq() {
	// fmt.Println("Try to broadcast RIP Req")
	node.RT.BroadcastRIPReq()
}

func (node *Node) HandleBroadcastRIPResp() {
	// fmt.Println("Try to broadcast RIP Resp")
	node.RT.BroadcastRIPResp()
}

// ***********************************************************************************
// Handle Receive Packet
func (node *Node) HandleReceivePacket(bytes []byte, destAddr string) {
	// check if  match can any port and the port is still alive
	// fmt.Println("Receive a packet")
	if !node.RT.CheckPktValidity(bytes, destAddr) {
		return
	}
	h, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln("Parse Header", err)
	}
	b := proto.UnmarshalRIPBody(bytes[20:])
	// HandleRIPResp or HandleTest
	switch h.Protocol {
	case 200:
		if b.Command == 1 {
			// fmt.Printf("Receive a RIP Req Packet from %v\n", destAddr)
			node.RT.HandleRIPReq(h.Src.String())
		} else {
			// fmt.Printf("Receive a RIP Resp Packet from %v\n", destAddr)
			node.RT.HandleRIPResp(bytes)
		}
	case 0:
		// fmt.Printf("Receive a TEST Packet from %v\n", destAddr)
		node.HandleTest(bytes)
	}
}

func (node *Node) HandleRIPResp(bytes []byte) {
	node.RT.HandleRIPResp(bytes)
}

// ***********************************************************************************
// Handle Test Packet
func (node *Node) HandleTest(bytes []byte) {
	node.RT.ForwardTestPkt(bytes)
}

// ***********************************************************************************
// Handle Expired Route
func (node *Node) HandleRouteEx(destIP string) {
	node.RT.CheckRouteEx(destIP)
}
