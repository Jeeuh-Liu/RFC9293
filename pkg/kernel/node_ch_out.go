package kernel

import (
	"fmt"
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
		case segment := <-node.segSendChan:
			// fmt.Println(segment)
			node.sendOutSegment(segment)
		}
	}
}
