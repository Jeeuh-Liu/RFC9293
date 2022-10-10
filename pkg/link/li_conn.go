package link

import (
	"fmt"
	"log"
	"net"
	"tcpip/pkg/proto"
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

// *****************************************************************************
// Read bytes from link
func (li *LinkInterface) ServeLink() {
	for {
		bytes := make([]byte, 1400)
		bnum, sourceAddr, err := li.LinkConn.ReadFromUDP(bytes)
		if err != nil {
			log.Fatalln(err)
		}
		// fmt.Printf("Receive %v bytes\n", bnum)

		// fmt.Printf("Receive bytes from %v\n", sourceAddr.String())
		// if the sourceAddr does not belong to this link, abandon it directly
		destAddr := sourceAddr.String()
		if destAddr != li.MACRemote {
			fmt.Printf("%v Not match %v", destAddr, li.MACRemote)
			continue
		}

		// send a CLI to handle packet
		cli := proto.NewCLI(proto.TypeHandlePacket, 0, bytes[:bnum], destAddr)
		li.NodeChan <- cli
	}
}

// ****************************************************************************
// Send bytes through link
func (li *LinkInterface) SendPacket(packetBytes []byte) {
	if li.Status != "up" {
		return
	}
	// fmt.Printf("Link try to send a RIP to %v\n", li.MACRemote)
	remoteAddr, err := net.ResolveUDPAddr("udp", li.MACRemote)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = li.LinkConn.WriteToUDP(packetBytes, remoteAddr)
	if err != nil {
		log.Fatalln("sendRIP", err)
	}
	// fmt.Printf("Send %v bytes\n", bnum)
}
