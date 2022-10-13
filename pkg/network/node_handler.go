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
	// add routes of local IP back to route table
	node.LocalIPSet[li.IPLocal] = true
	node.DestIP2Route[li.IPLocal] = route
	// we do not need to handle remote routes manually
	// change status of link
	node.ID2Interface[uint8(id)].OpenRemoteLink()
}

func (node *Node) HandleSetDown(id uint8) {
	li := node.ID2Interface[id]
	// delete the routes of local IP from route table
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
	// change status of link
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

func (node *Node) HandleSendPacket(destIP string, protoID int, msg string) {
	if route, ok := node.DestIP2Route[destIP]; ok {
		// Choose the link whose IPRemote == nextIP to send
		for _, li := range node.ID2Interface {
			if li.IPRemote == route.Next {
				fmt.Printf("Try to send a packet from %v to %v\n", li.IPLocal, destIP)
				test := proto.NewPktTest(li.IPLocal, destIP, msg, 16)
				bytes := test.Marshal()
				li.SendPacket(bytes)
				return
			}
		}
	}
	fmt.Println("destIP does not exist")
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
		rip := proto.NewPktRIP(li.IPLocal, li.IPRemote, 1, entries)
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
		rip := proto.NewPktRIP(li.IPLocal, li.IPRemote, 2, entries)

		bytes := rip.Marshal()
		li.SendPacket(bytes)
	}
}

// ***********************************************************************************
// Handle Receive Packet
func (node *Node) HandleReceivePacket(bytes []byte, destAddr string) {
	// check if  match can any port and the port is still alive
	// fmt.Println("Receive a packet")
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
		fmt.Printf("%v does not match and be alive\n", destAddr)
		return
	}
	// check length of bytes
	if len(bytes) < 20 {
		// fmt.Println(len(bytes))
		return
	}
	h, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln("Parse Header", err)
	}
	if h.TotalLen != len(bytes) {
		// fmt.Println(h.TotalLen, len(bytes))
		return
	}
	// Check sum
	// fmt.Println(bytes)
	// curChecksum := int(proto.ComputeChecksum(bytes[:20]))
	// if h.Checksum != curChecksum {
	// 	fmt.Println("Should be:", h.Checksum, ", Current:", curChecksum)
	// 	return
	// }
	// HandleRIPResp or HandleTest
	switch h.Protocol {
	case 200:
		// fmt.Printf("Receive a RIP Packet from %v\n", destAddr)
		node.HandleRIPResp(bytes)
	case 0:
		// fmt.Printf("Receive a TEST Packet from %v\n", destAddr)
		node.HandleTest(bytes)
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
		// 1. Expiration time
		// (1) If the destIP is local IP of 1 interface, it will not expire
		destIP := ipv4Num2str(entry.Address)
		if _, ok := node.LocalIPSet[destIP]; ok {
			continue
		}
		// (2) If the destIP is remote IP, Update new Expiration time
		node.RemoteDest2ExTime[destIP] = time.Now().Add(12 * time.Second)
		go node.SendExTimeCLI(destIP)
		// fmt.Println(rip.Header.Src)

		// 2. Route
		oldCost := node.DestIP2Route[destIP].Cost
		oldNextAddr := node.DestIP2Route[destIP].Dest
		newCost := entry.Cost + 1
		newNextAddr := rip.Header.Src.String()
		newRoute := NewRoute(destIP, newNextAddr, newCost)
		// (1) If no existing , update
		if _, ok := node.RemoteDestIP2Cost[destIP]; !ok {
			node.UpdateRoutes(newRoute, destIP)
		}
		// (2) If newCost < oldCost, update
		if newCost < oldCost {
			node.UpdateRoutes(newRoute, destIP)
		}
		// (3) If newCost > oldCost and newNextAddr == oldNextAddr, update
		if newCost > oldCost && newNextAddr == oldNextAddr {
			node.UpdateRoutes(newRoute, destIP)
		}
		// (4) If newCost > oldCost and newNextAddr != oldNextAddr, ignore
		if newCost > oldCost && newNextAddr != oldNextAddr {
			continue
		}
		// (5) If newCost == oldCost, ignore
		if newCost == oldCost {
			continue
		}
	}
}

func (node *Node) UpdateRoutes(newRoute Route, destIP string) {
	// update routes
	node.DestIP2Route[destIP] = newRoute
	// update the metadata
	node.RemoteDestIP2Cost[destIP] = newRoute.Cost
	node.RemoteDestIP2SrcIP[destIP] = newRoute.Next
	// Broadcast RIP Resp because of Triggered Updates
	entry := proto.NewEntry(newRoute.Cost, newRoute.Dest)
	node.BroadcastRIPRespTU(entry)
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

// ***********************************************************************************
// Handle Test Packet
func (node *Node) HandleTest(bytes []byte) {
	test := proto.UnmarshalPktTest(bytes)
	srcIP := test.Header.Src.String()
	destIP := test.Header.Dst.String()
	msg := string(test.Body)
	ttl := test.Header.TTL
	// 1. Validity
	// (1) Is checksum in header valid?
	// h, err := ipv4.ParseHeader(bytes[:20])
	// if err != nil {
	// 	log.Fatalln("Parse Header", err)
	// }
	// curChecksum := int(proto.ComputeChecksum(bytes[:20]))
	// if h.Checksum != curChecksum {
	// 	fmt.Println("Should be:", h.Checksum, ", Current:", curChecksum)
	// 	return
	// }
	// (2) Is ttl == 0 ?
	if ttl == 0 {
		return
	}

	// 2. Forwarding
	// (1) Does this packet belong to me?
	if _, ok := node.LocalIPSet[destIP]; ok {
		fmt.Printf("---Node received packet!---\n")
		fmt.Printf("        source IP      : %v\n", srcIP)
		fmt.Printf("        destination IP : %v\n", destIP)
		fmt.Printf("        protocol       : %v\n", 0)
		fmt.Printf("        payload length : %v\n", len(msg))
		fmt.Printf("        payload        : %v\n", msg)
		fmt.Printf("----------------------------\n")
		return
	}
	// (2) Does packet match any IF in the forwarding table?
	if route, ok := node.DestIP2Route[destIP]; ok {
		// Choose the link whose IPRemote == nextIP to send
		for _, li := range node.ID2Interface {
			if li.IPRemote == route.Next {
				fmt.Printf("Try to send a packet from %v to %v\n", li.IPLocal, destIP)
				test := proto.NewPktTest(srcIP, destIP, msg, ttl-1)
				bytes := test.Marshal()
				li.SendPacket(bytes)
			}
		}
	}
	// (3) Does the router have next hop?
	fmt.Println("destIP does not exist")
}
