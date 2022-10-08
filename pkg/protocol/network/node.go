package network

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"tcpip/pkg/protocol/link"
)

// The driver program

type Node struct {
	Mu           sync.Mutex
	ID2Interface map[uint8]*link.LinkInterface
	NodeCLIChan  chan *CLI
	// Routers
	Routes []Route
	// Local MAC addr and UDPConn
	MACLocal  string
	LocalConn *net.UDPConn
	// RIP metadata
	// check min_cost
	RemoteDestIP2Cost map[string]uint32
	// check Split Horizon with Poisoned Reverse
	RemoteDestIP2SrcIP map[string]string
}

func (node *Node) Make(args []string) {
	// Initialize NodeCLIChan
	node.NodeCLIChan = make(chan *CLI)
	// Initialize ID2Interface
	node.ID2Interface = make(map[uint8]*link.LinkInterface)
	inx := args[1]
	f, err := os.Open(inx)
	if err != nil {
		log.Fatalln(err)
	}
	r := bufio.NewReader(f)

	id := uint8(0)
	for {
		bytes, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln(err)
		}
		eles := strings.Split(string(bytes), " ")
		li := &link.LinkInterface{}
		if len(eles) == 2 {
			udpPortLocal := eles[1]
			node.MACLocal = ToIPColonAddr("localhost", udpPortLocal)
			fmt.Println("MACLocal is", node.MACLocal)
			continue
		}
		//id, udpIp, udpPortRemote, ipLocal, ipRemote
		li.Make(id, eles[0], eles[1], eles[2], eles[3])
		fmt.Printf("%v: %v\n", id, eles[2])
		node.ID2Interface[id] = li
		id++
	}
	// fmt.Println(node)
	// Initialize Routes: each interface to itself
	node.Routes = []Route{}
	for _, li := range node.ID2Interface {
		route := Route{
			Dest: li.IPLocal,
			Next: li.IPLocal,
			Cost: 0,
		}
		node.Routes = append(node.Routes, route)
	}
	// initialize map remote2cost
	node.RemoteDestIP2Cost = map[string]uint32{}
	// initialize map remoteDest2src
	node.RemoteDestIP2SrcIP = map[string]string{}
	// Receive UDP
	go node.ServeLocalLink()
	// Receive CLI
	go node.ScanClI()
	// Broadcast RIP periodically
	go node.RIPDaemon()
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
