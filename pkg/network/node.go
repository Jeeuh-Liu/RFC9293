package network

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"tcpip/pkg/link"
	"tcpip/pkg/proto"
	"time"
)

// The driver program

type Node struct {
	ID2Interface map[uint8]*link.LinkInterface // Store interfaces and facilitate down and up
	// Channel
	NodeCLIChan   chan *proto.NodeCLI   // Receive CLI from user
	NodeBCChan    chan *proto.NodeBC    // Broadcast RIP
	NodeExChan    chan *proto.NodeEx    // Handle expiration of route
	NodePktOpChan chan *proto.NodePktOp // Receive msg from link interface
	// Routers
	DestIP2Route map[string]Route //store routes and facilitate finding target route for Test Packet
	// RIP metadata
	// check min_cost
	LocalIPSet         map[string]bool      // store all local IP in this node to facilitate test packet checking
	RemoteDestIP2Cost  map[string]uint32    // ensure min cost of new route
	RemoteDestIP2SrcIP map[string]string    // check Split Horizon with Poisoned Reverse to set cost = 16
	RemoteDest2ExTime  map[string]time.Time // Check Expiration time of a new route
}

func (node *Node) Make(args []string) {
	// Initialize NodeCLIChan, we can set to bigger to avoid some deadlock
	node.NodeCLIChan = make(chan *proto.NodeCLI)
	// Initialize NodeOpChan
	node.NodeBCChan = make(chan *proto.NodeBC)
	node.NodeExChan = make(chan *proto.NodeEx)
	node.NodePktOpChan = make(chan *proto.NodePktOp)
	// Initialize ID2Interface
	node.ID2Interface = make(map[uint8]*link.LinkInterface)
	inx := args[1]
	f, err := os.Open(inx)
	if err != nil {
		log.Fatalln(err)
	}
	r := bufio.NewReader(f)

	id := uint8(0)

	var udpPortLocal string
	// Open linkConn with first line
	bytes, _, err := r.ReadLine()
	if err != nil {
		log.Fatalln("ReadFirstLine", err)
	}
	eles := strings.Split(string(bytes), " ")
	localAddr, err := net.ResolveUDPAddr("udp", ToIPColonAddr(eles[0], eles[1]))
	if err != nil {
		log.Fatalln("Resolve UDPAddr", err)
	}
	linkConn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		log.Fatalln("ListenUDP", err)
	}
	// Initialize link Interface
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
		// elements: udpIp, udpPortRemote, ipLocal, ipRemote
		//li.Make(udpIp, udpPortRemote, ipLocal, ipRemote, id, udpPortLocal)
		li.Make(eles[0], eles[1], eles[2], eles[3], id, udpPortLocal, linkConn, node.NodePktOpChan)
		fmt.Printf("%v: %v\n", id, eles[2])
		node.ID2Interface[id] = li
		id++
	}
	// fmt.Println(node)
	// Initialize Routes: each interface to itself
	// initialize local IP set
	node.LocalIPSet = map[string]bool{}
	node.DestIP2Route = map[string]Route{}
	for _, li := range node.ID2Interface {
		route := Route{
			Dest: li.IPLocal,
			Next: li.IPLocal,
			Cost: 0,
		}
		node.DestIP2Route[route.Dest] = route
		node.LocalIPSet[li.IPLocal] = true
	}
	// initialize map remote2cost
	node.RemoteDestIP2Cost = map[string]uint32{}
	// initialize map remoteDest2src
	node.RemoteDestIP2SrcIP = map[string]string{}
	// initialize map remoteDest2exTime
	node.RemoteDest2ExTime = map[string]time.Time{}
	// Receive CLI
	go node.ScanClI()
	// Broadcast RIP Request once
	go node.RIPReqDaemon()
	// Broadcast RIP Resp periodically
	go node.RIPRespDaemon()
}

func ToIPColonAddr(udpIp, udpPort string) string {
	return fmt.Sprintf("%v:%v", udpIp, udpPort)
}
