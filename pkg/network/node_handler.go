package network

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"tcpip/pkg/proto"
	"time"

	"golang.org/x/net/ipv4"
)

// ***********************************************************************************
// Handle CLI
func (node *Node) HandlePrintInterfaces() {
	fmt.Println("id    state        local        remote        port")
	for id, li := range node.ID2Interface {
		port := strings.Split(li.MACRemote, ":")[1]
		fmt.Printf("%v      %v         %v     %v      %v\n", id, li.Status, li.IPLocal, li.IPRemote, port)
	}
}

func (node *Node) HandleSetUp(id uint8) {
	li := node.ID2Interface[id]
	route := NewRoute(li.IPLocal, li.IPLocal, 0)
	node.LocalIPSet[li.IPLocal] = true
	node.DestIP2Route[li.IPLocal] = route
	node.ID2Interface[uint8(id)].OpenRemoteLink()
}

func (node *Node) HandleSetDown(id uint8) {
	li := node.ID2Interface[id]
	// delete the local routes
	delete(node.LocalIPSet, li.IPLocal)
	delete(node.DestIP2Route, li.IPLocal)
	// if a remote destIP needs to use this link, delete corresponding its route and metadata
	for destIP, route := range node.DestIP2Route {
		if route.Next == li.IPRemote {
			delete(node.DestIP2Route, destIP)
			delete(node.RemoteDest2ExTime, destIP)
			delete(node.RemoteDestIP2Cost, destIP)
			delete(node.RemoteDestIP2SrcIP, destIP)
		}
	}
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

// ***********************************************************************************
// Handle BroadcastRIP
func (node *Node) HandleBroadcastRIPReq() {
	// fmt.Println("Try to broadcast RIP Req")
	for _, li := range node.ID2Interface {
		if !li.IsUp() {
			continue
		}
		entries := []proto.Entry{}
		rip := proto.NewRIP(li.IPLocal, li.IPRemote, 1, entries)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

func (node *Node) HandleBroadcastRIPResp() {
	// fmt.Println("Try to broadcast RIP Resp")
	for _, li := range node.ID2Interface {
		if !li.IsUp() {
			continue
		}
		entries := []proto.Entry{}
		// For RIP resp, we need to load all valid entries into RIP body
		for _, route := range node.DestIP2Route {
			// if route.next == src of route.dest -> ignore this route entry
			entry := proto.NewEntry(route.Cost, route.Dest)
			if srcIP, ok := node.RemoteDestIP2SrcIP[route.Dest]; ok && srcIP == li.IPRemote {
				entry.Cost = 16
			}
			entries = append(entries, entry)
			// fmt.Println(entries)
		}
		rip := proto.NewRIP(li.IPLocal, li.IPRemote, 2, entries)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

// ***********************************************************************************
// Handle Packet
func (node *Node) HandlePacket(bytes []byte, destAddr string) {
	// checksum

	// check if  match can any port and the port is still alive
	canMatch := false
	isAlive := false
	for _, li := range node.ID2Interface {
		if destAddr == li.MACRemote {
			canMatch = true
			if li.IsUp() {
				isAlive = true
			}
		}
	}
	if !canMatch || !isAlive {
		// fmt.Printf("%v does not match and be alive\n", destAddr)
		return
	}
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln("Parse Header", err)
	}
	switch header.Protocol {
	case 200:
		// fmt.Printf("Receive a RIP Packet from %v\n", destAddr)
		node.HandleRIPResp(bytes)
	case 0:
		// fmt.Printf("Receive a TEST Packet from %v\n", destAddr)

	}
}

func (node *Node) HandleRIPResp(bytes []byte) {
	rip := proto.UnmarshalRIPResp(bytes)
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
		// If the destIP is local IP of 1 interface, it will not expire
		destIP := ipv4Num2str(entry.Address)
		if _, ok := node.LocalIPSet[destIP]; ok {
			continue
		}
		// fmt.Printf("Receive a dest addr %v\n", destIP)
		// Update new Expiration time
		node.RemoteDest2ExTime[destIP] = time.Now().Add(12 * time.Second)
		go node.SendExTimeCLI(destIP)
		// fmt.Println(rip.Header.Src)
		// Min Cost
		// if the dest addr exists in destAddr2Cost and new cost is bigger, ignore
		newCost := entry.Cost + 1
		// fmt.Printf("newCost is %v\n", newCost)
		if cost, ok := node.RemoteDestIP2Cost[destIP]; ok && newCost >= cost {
			// fmt.Printf("oldCost is %v\n", cost)
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
		// Broadcast RIP Resp because of Triggered Updates
		// proto.NewEntry(newRoute.Cost, newRoute.Dest)
		// node.BroadcastRIPRespTU(entry)
	}
}

func (node *Node) BroadcastRIPRespTU(entity proto.Entry) {
	for _, li := range node.ID2Interface {
		if !li.IsUp() {
			continue
		}
		entries := []proto.Entry{}
		entries = append(entries, entity)
		rip := proto.NewRIP(li.IPLocal, li.IPRemote, 2, entries)
		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

// ***********************************************************************************
// Handle Expired Route
func (node *Node) HandleRouteEx(destIP string) {
	if time.Now().After(node.RemoteDest2ExTime[destIP]) {
		delete(node.DestIP2Route, destIP)
		delete(node.RemoteDest2ExTime, destIP)
		delete(node.RemoteDestIP2Cost, destIP)
		delete(node.RemoteDestIP2SrcIP, destIP)
	}
}

func ipv4Num2str(addr uint32) string {
	mask := 1<<8 - 1
	res := strconv.Itoa(int(addr) & mask)
	addr >>= 8
	for i := 0; i < 3; i++ {
		res = strconv.Itoa(int(addr)&mask) + "." + res
		addr >>= 8
	}
	return res
}