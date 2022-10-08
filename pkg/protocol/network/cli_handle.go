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

// BroadcastRIP
func (node *Node) BroadcastRIP() {
	// fmt.Println("Try to broadcast RIP")
	for _, li := range node.ID2Interface {
		rip := node.NewRIP(li.IPLocal, li.IPRemote)
		bytes := rip.Marshal()
		li.SendRIP(bytes)
	}
}

// HandleRIP
func (node *Node) HandleRIP(bytes []byte) {
	rip := UnmarshalRIP(bytes)
	num_entries := rip.Body.num_entries
	for i := 0; i < int(num_entries); i++ {
		entry := rip.Body.entries[i]
		newCost := entry.cost + 1
		// fmt.Printf("newCost is %v\n", newCost)
		destAddr := ipv4Num2str(entry.address)
		// fmt.Printf("Receive a dest addr %v\n", destAddr)
		// if the dest addr exists in destAddr2Cost and new cost is bigger, ignore
		if cost, ok := node.RemoteDestIP2Cost[destAddr]; ok && newCost >= cost {
			continue
		}
		fmt.Println(rip.Header.Src)
		nextAddr := netIP2str(rip.Header.Src)
		// fmt.Printf("nextAddr is %v\n", nextAddr)
		newRoute := NewRoute(destAddr, nextAddr, newCost)
		// fmt.Println(newRoute)
		node.Routes = append(node.Routes, newRoute)
		// update the metadata
		node.RemoteDestIP2Cost[destAddr] = newCost
		node.RemoteDestIP2SrcIP[destAddr] = nextAddr
	}
}
