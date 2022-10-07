package network

import (
	"fmt"
	"os"
	"strings"
)

// Handle CLI
func (node *Node) PrintInterfaces() {
	fmt.Println("id    state        local        remote        port")
	for id, li := range node.ID2Interface {
		port := strings.Split(li.MACRemote, ":")[1]
		fmt.Printf("%v      %v         %v     %v      %v\n", id, li.Status, li.IPLocal, li.IPRemote, port)
	}
}

func (node *Node) SetUp(id uint8) {
	node.ID2Interface[uint8(id)].OpenRemoteLink()
}

func (node *Node) SetDown(id uint8) {
	node.ID2Interface[uint8(id)].CloseRemoteLink()
}

func (node *Node) Quit() {
	// for _, li := range node.ID2Interface {
	// 	li.CloseLink()
	// }
	os.Exit(0)
}

func (node *Node) PrintRoutes() {
	fmt.Println("    dest        	next        cost")
	for _, r := range node.Routes {
		fmt.Printf("    %v         %v         %v\n", r.Dest, r.Next, r.Cost)
	}
}

// HandleRIP
func (node *Node) HandleRIP(bytes []byte) {

}
