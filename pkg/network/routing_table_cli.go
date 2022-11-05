package network

import (
	"fmt"
	"log"
	"os"
	"strings"
	"tcpip/pkg/proto"

	"github.com/google/netstack/tcpip/header"
)

func (rt *RoutingTable) PrintInterfaces() {
	fmt.Println("id    state        local        remote        port")
	for id, li := range rt.ID2Interface {
		port := strings.Split(li.MACRemote, ":")[1]
		fmt.Printf("%v      %v         %v     %v      %v\n", id, li.Status, li.IPLocal, li.IPRemote, port)
	}
}

func (rt *RoutingTable) PrintInterfacesToFile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	header := "id    state        local        remote        port\n"
	_, err = f.WriteString(header)
	if err != nil {
		log.Println(err)
	}
	for id, li := range rt.ID2Interface {
		port := strings.Split(li.MACRemote, ":")[1]
		line := fmt.Sprintf("%v      %v         %v     %v      %v\n", id, li.Status, li.IPLocal, li.IPRemote, port)
		_, err = f.WriteString(line)
		if err != nil {
			log.Println(err)
		}
	}
}

func (rt *RoutingTable) SetUp(id uint8) {
	li := rt.ID2Interface[id]
	// set routes of local IP back to 0
	newRoute := NewRoute(li.IPLocal, li.IPLocal, 0)
	rt.UpdateRoutesAndBroadcastTU(newRoute, li.IPLocal)
	// change status of link
	rt.ID2Interface[uint8(id)].OpenRemoteLink()
}

func (rt *RoutingTable) SetDown(id uint8) {
	li := rt.ID2Interface[id]
	// set routes of local IP to infinity
	newRoute := NewRoute(li.IPLocal, li.IPLocal, 16)
	rt.UpdateRoutesAndBroadcastTU(newRoute, li.IPLocal)
	// if a remote destIP needs to use this link, delete corresponding its route and metadata
	for destIP, route := range rt.DestIP2Route {
		if route.Next == li.IPRemote {
			// regard this destIP as expired
			newRoute.Cost = 16
			rt.UpdateRoutesAndBroadcastTU(newRoute, destIP)
		}
	}
	// change status of link
	rt.ID2Interface[uint8(id)].CloseRemoteLink()
}

func (rt *RoutingTable) PrintRoutes() {
	fmt.Println("    dest        	next        cost")
	for _, r := range rt.DestIP2Route {
		if r.Cost != 16 {
			fmt.Printf("    %v         %v         %v\n", r.Dest, r.Next, r.Cost)
		}
	}
}

func (rt *RoutingTable) PrintRoutesToFile(filename string) {
	f, err := os.Create(filename)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	header := "    dest        	next        cost\n"
	_, err = f.WriteString(header)
	if err != nil {
		log.Println(err)
	}
	for _, r := range rt.DestIP2Route {
		if r.Cost != 16 {
			line := fmt.Sprintf("    %v         %v         %v\n", r.Dest, r.Next, r.Cost)
			_, err = f.WriteString(line)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (rt *RoutingTable) SendPacket(destIP string, protoID int, msg string) {
	ttl := 16
	if route, ok := rt.DestIP2Route[destIP]; ok && route.Cost < 16 {
		// check if route.cost == inf => unreachable
		// Choose the link whose IPRemote == nextIP to send
		for _, li := range rt.ID2Interface {
			if li.IPRemote == route.Next {
				// fmt.Printf("Try to send a packet from %v to %v\n", li.IPLocal, destIP)
				test := proto.NewPktTest(li.IPLocal, destIP, msg, ttl-1)
				if protoID == 6 {
					test.Header.Protocol = 6
					test.Header.Flags = header.IPv4FlagDontFragment
				}
				test.Header.Checksum = 0
				headerBytes, err := test.Header.Marshal()
				if err != nil {
					log.Fatalln("Error marshalling header:  ", err)
				}
				test.Header.Checksum = int(proto.ComputeChecksum(headerBytes))

				bytes := test.Marshal()
				// proto.PrintHex(bytes)
				li.SendPacket(bytes)
				return
			}
		}
	}
	fmt.Println("destIP does not exist")
}
