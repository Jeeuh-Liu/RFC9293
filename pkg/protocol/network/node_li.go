package network

import (
	"log"
	"net"
	"time"

	"golang.org/x/net/ipv4"
)

// Broadcast RIP through LinkInterface
func (node *Node) RIPDaemon() {
	for {
		cli := NewCLI(RIPBroadcast, 0, []byte{}, "")
		node.NodeCLIChan <- cli
		time.Sleep(5 * time.Second)
	}
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

		Header, err := ipv4.ParseHeader(bytes[:20])
		if err != nil {
			log.Fatalln(err)
		}
		switch Header.Protocol {
		case 200:
			CLI := NewCLI(RIPHandle, 0, bytes, "")
			node.NodeCLIChan <- CLI
		case 0:
			// fmt.Println("Receive a TEST")
			// CLI := NewCLI(RIP, 0, bytes)
			// node.NodeCLIChan <- CLI
		}
	}
}
