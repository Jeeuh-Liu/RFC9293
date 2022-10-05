package ip

import (
	"fmt"
	"os"
	"strings"
)

func (node *Node) PrintInterfaces() {
	fmt.Println("id    state        local        remote        port")
	for id, li := range node.ID2Interface {
		port := strings.Split(li.Addr, ":")[1]
		fmt.Printf("%v      %v         %v     %v      %v\n", id, li.Status, li.IpLocal, li.IpRemote, port)
	}
}

func (node *Node) SetUp(id uint8) {
	node.ID2Interface[uint8(id)].OpenLink()
}

func (node *Node) SetDown(id uint8) {
	node.ID2Interface[uint8(id)].CloseLink()
}

func (node *Node) Quit() {
	// for _, li := range node.ID2Interface {
	// 	li.CloseLink()
	// }
	os.Exit(0)
}
