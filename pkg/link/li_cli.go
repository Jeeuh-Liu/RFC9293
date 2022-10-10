package link

import (
	"fmt"
	"log"
	"net"
)

// Handle commands to open and close of link
func (li *LinkInterface) OpenRemoteLink() {
	if li.Status == "up" {
		fmt.Printf("interface %v is already up\n", li.ID)
		return
	}
	// fmt.Println(li.Addr, li.IpLocal, li.IpRemote)
	localAddr, err := net.ResolveUDPAddr("udp", li.MACLocal)
	if err != nil {
		log.Fatalln(err)
	}
	linkConn, err := net.ListenUDP("udp", localAddr)
	li.LinkConn = linkConn
	if err != nil {
		log.Fatalln("Open LinkConn", err)
	}
	li.Status = "up"
	fmt.Printf("interface %v is now enabled, Dial to udp %v\n", li.ID, li.MACRemote)
	go li.ServeLink()
}

func (li *LinkInterface) CloseRemoteLink() {
	if li.Status == "down" {
		fmt.Printf("interface %v is already down\n", li.ID)
		return
	}
	li.LinkConn.Close()
	li.Status = "dn"
	fmt.Printf("interface %v is now disabled\n", li.ID)
}

func (li *LinkInterface) ServeLink() {
	for {
		bytes := make([]byte, 1400)
		bnum, sourceAddr, err := li.LinkConn.ReadFromUDP(bytes)
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Printf("Receive %v bytes\n", bnum)

		// fmt.Printf("Receive bytes from %v\n", sourceAddr.String())
		// if the sourceAddr does not belong to this link, abandon it directly
		if sourceAddr.String() != li.MACRemote {
			fmt.Printf("%v Not match %v", sourceAddr.String(), li.MACRemote)
			continue
		}

		// rip := UnmarshalRIPResp(bytes)
		// if err != nil {
		// 	log.Fatalln(err)
		// }
		// switch rip.Header.Protocol {
		// case 200:
		// 	switch rip.Body.Command {
		// 	case 1:
		// 		fmt.Println("Receive a RIP Request")
		// 		// CLI := NewCLI(RIPReqHandle, 0, bytes, ")
		// 		// node.NodeCLIChan <- LI
		// 	case 2:
		// 		fmt.Println("Receive a RIP Response")
		// 		// CLI := NewCLI(RIPRespHandle, 0, bytes, "")
		// 		// node.NodeCLIChan <- CLI
		// 	}
		// case 0:
		// 	// fmt.Println("Receive a TEST")
		// }
	}
}
