package network

import (
	"fmt"
	"log"
	"net"
	"time"
)

// Broadcast RIP through LinkInterface
func (node *Node) RIPRespDaemon() {
	for {
		cli := NewCLI(RIPRespBroadcast, 0, []byte{}, "")
		node.NodeCLIChan <- cli
		time.Sleep(5 * time.Second)
	}
}

func (node *Node) RIPReqDaemon() {
	cli := NewCLI(RIPReqBroadcast, 0, []byte{}, "")
	node.NodeCLIChan <- cli
}

// Receive bytes though link
func (node *Node) ServeLocalLink() {
	addr, err := net.ResolveUDPAddr("udp", node.MACLocal)
	if err != nil {
		log.Fatalln(err)
	}
	// fmt.Println(li.Addr, li.IpLocal, li.IpRemote)
	node.LocalConn, err = net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		bytes := make([]byte, 1400)
		// bnum, err := node.LocalConn.Read(bytes)
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// fmt.Printf("Receive %v bytes\n", bnum)
		// fmt.Println(node.LocalConn.RemoteAddr().String())
		node.LocalConn.Read(bytes)

		rip := UnmarshalRIPResp(bytes)
		if err != nil {
			log.Fatalln(err)
		}
		switch rip.Header.Protocol {
		case 200:
			switch rip.Body.Command {
			case 1:
				fmt.Println("Receive a RIP Request")
				// CLI := NewCLI(RIPReqHandle, 0, bytes, ")
				// node.NodeCLIChan <- LI
			case 2:
				fmt.Println("Receive a RIP Response")
				CLI := NewCLI(RIPRespHandle, 0, bytes, "")
				node.NodeCLIChan <- CLI
			}
		case 0:
			// fmt.Println("Receive a TEST")
		}
	}
}
