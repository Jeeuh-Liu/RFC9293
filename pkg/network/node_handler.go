package network

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/net/ipv4"
)

// Handle CLI
// ***********************************************************************************
func (node *Node) HandlePrintInterfaces() {
	fmt.Println("id    state        local        remote        port")
	for id, li := range node.ID2Interface {
		port := strings.Split(node.MACLocal, ":")[1]
		fmt.Printf("%v      %v         %v     %v      %v\n", id, li.Status, li.IPLocal, li.IPRemote, port)
	}
}

func (node *Node) HandleSetUp(id uint8) {
	node.ID2Interface[uint8(id)].OpenRemoteLink()
}

func (node *Node) HandleSetDown(id uint8) {
	node.ID2Interface[uint8(id)].CloseRemoteLink()
}

func (node *Node) HandleQuit() {
	os.Exit(0)
}

func (node *Node) HandlePrintRoutes() {
	fmt.Println("    dest        	next        cost")
	for _, r := range node.DestIP2Route {
		fmt.Printf("    %v         %v         %v\n", r.Dest, r.Next, r.Cost)
	}
}

// Handle BroadcastRIP
// ***********************************************************************************
func (node *Node) HandleBroadcastRIPResp() {
	// fmt.Println("Try to broadcast RIP Resp")
	for _, li := range node.ID2Interface {
		rip := node.NewRIPResp(li)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

func (node *Node) HandleBroadcastRIPReq() {
	// fmt.Println("Try to broadcast RIP Req")
	for _, li := range node.ID2Interface {
		rip := node.NewRIPReq(li)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

// Handle Packet
// ***********************************************************************************
func (node *Node) HandlePacket(bytes []byte) {
	// checksum
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln("Parse Header", err)
	}
	switch header.Protocol {
	case 200:
		// fmt.Println("Receive a RIP packet")
		// CLI := proto.NewCLI(proto.TypeHandleRIPResp, 0, bytes, "")
		// node.NodeCLIChan <- CLI
		node.HandleRIPResp(bytes)
	case 0:
		// fmt.Println("Receive a TEST")

	}
}

func (node *Node) HandleRIPResp(bytes []byte) {
	rip := UnmarshalRIPResp(bytes)
	// fmt.Println(rip.Header.Src.String(), rip.Header.Dst.String())
	num_entries := rip.Body.Num_Entries
	// fmt.Println(num_entries)
	for i := 0; i < int(num_entries); i++ {
		entry := rip.Body.Entries[i]
		// fmt.Println(entry)
		// if entry.cost == 16, sending back -> ignore
		if entry.Cost == 16 {
			continue
		}
		// Expiration time
		destIP := ipv4Num2str(entry.Address)
		if _, ok := node.LocalIPSet[destIP]; ok {
			continue
		}
		// fmt.Printf("newCost is %v\n", newCost)
		// fmt.Printf("Receive a dest addr %v\n", destIP)
		node.RemoteDest2ExTime[destIP] = time.Now().Add(12 * time.Second)
		go node.SendExTimeCLI(destIP)
		// fmt.Println(rip.Header.Src)
		// Min Cost
		// if the dest addr exists in destAddr2Cost and new cost is bigger, ignore
		newCost := entry.Cost + 1
		if cost, ok := node.RemoteDestIP2Cost[destIP]; ok && newCost >= cost {
			continue
		}
		nextAddr := rip.Header.Src.String()
		// fmt.Printf("nextAddr is %v\n", nextAddr)
		newRoute := NewRoute(destIP, nextAddr, newCost)
		// fmt.Println("newRoute:", newRoute)
		node.DestIP2Route[destIP] = newRoute
		// fmt.Println(node.DestIP2Route)
		// update the metadata
		node.RemoteDestIP2Cost[destIP] = newCost
		node.RemoteDestIP2SrcIP[destIP] = nextAddr
	}
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
