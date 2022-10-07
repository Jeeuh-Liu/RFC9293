package network

import (
	"fmt"
	"log"
	"net"

	"golang.org/x/net/ipv4"
)

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
		node.LocalConn.Read(bytes)

		Header, err := ipv4.ParseHeader(bytes[:20])
		if err != nil {
			log.Fatalln(err)
		}
		switch Header.Protocol {
		case 200:
			fmt.Println("Receive a RIP")
			CLI := NewCLI(RIP, 0, bytes[20:])
			node.NodeCLIChan <- CLI
		case 0:
			fmt.Println("Receive a TEST")
			CLI := NewCLI(RIP, 0, bytes[20:])
			node.NodeCLIChan <- CLI
		}
	}
}
