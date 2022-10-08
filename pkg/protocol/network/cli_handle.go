package network

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Handle CLI
// ***********************************************************************************
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
	for _, r := range node.DestIP2Route {
		fmt.Printf("    %v         %v         %v\n", r.Dest, r.Next, r.Cost)
	}
}

// BroadcastRIP
// ***********************************************************************************
func (node *Node) BroadcastRIP() {
	// fmt.Println("Try to broadcast RIP")
	for _, li := range node.ID2Interface {
		rip := node.NewRIP(li.IPLocal, li.IPRemote)
		bytes := rip.Marshal()
		li.SendRIP(bytes)
	}
}

// HandleRIP
// ***********************************************************************************
func (node *Node) HandleRIP(bytes []byte) {
	rip := UnmarshalRIP(bytes)
	num_entries := rip.Body.num_entries
	for i := 0; i < int(num_entries); i++ {
		entry := rip.Body.entries[i]
		// Expiration time
		destIP := ipv4Num2str(entry.address)
		// fmt.Printf("newCost is %v\n", newCost)
		// fmt.Printf("Receive a dest addr %v\n", destAddr)
		node.RemoteDest2ExTime[destIP] = time.Now().Add(12 * time.Second)
		go node.SendExTimeCLI(destIP)
		// Min Cost
		// if the dest addr exists in destAddr2Cost and new cost is bigger, ignore
		newCost := entry.cost + 1
		if cost, ok := node.RemoteDestIP2Cost[destIP]; ok && newCost >= cost {
			continue
		}
		fmt.Println(rip.Header.Src)
		nextAddr := netIP2str(rip.Header.Src)
		// fmt.Printf("nextAddr is %v\n", nextAddr)
		newRoute := NewRoute(destIP, nextAddr, newCost)
		// fmt.Println(newRoute)
		// node.Routes = append(node.Routes, newRoute)
		node.DestIP2Route[destIP] = newRoute
		// update the metadata
		node.RemoteDestIP2Cost[destIP] = newCost
		node.RemoteDestIP2SrcIP[destIP] = nextAddr
	}
}

func (node *Node) SendExTimeCLI(destIP string) {
	// sleep 13 second and check whether the time expires
	time.Sleep(13 * time.Second)
	cli := NewCLI(RouteEx, 0, []byte{}, destIP)
	node.NodeCLIChan <- cli
}

// Handle Expired Route
func (node *Node) HandleRouteEx(destIP string) {
	if time.Now().After(node.RemoteDest2ExTime[destIP]) {
		delete(node.RemoteDest2ExTime, destIP)
		delete(node.DestIP2Route, destIP)
		delete(node.RemoteDestIP2Cost, destIP)
		delete(node.RemoteDestIP2SrcIP, destIP)
	}
}
