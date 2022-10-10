package link

import (
	"fmt"
	"log"
	"net"
	"tcpip/pkg/proto"
)

// Open and close of link
// *****************************************************************************
func (li *LinkInterface) OpenRemoteLink() {
	li.Mu.Lock()
	defer li.Mu.Unlock()
	if li.Status == "up" {
		fmt.Printf("interface %v is already up\n", li.ID)
		return
	}
	li.Status = "up"
	fmt.Printf("interface %v is now enabled, Dial to udp %v\n", li.ID, li.MACRemote)
}

func (li *LinkInterface) CloseRemoteLink() {
	li.Mu.Lock()
	defer li.Mu.Unlock()
	if li.Status == "dn" {
		fmt.Printf("interface %v is already down\n", li.ID)
		return
	}
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
			// if the connection close, stop this goroutine
			// fmt.Println("The linkConn is closed")
			return
		}
		// fmt.Printf("Receive bytes from %v\n", sourceAddr.String())
		// if the sourceAddr does not belong to this link, abandon it directly
		destAddr := sourceAddr.String()
		// send a CLI to handle packet
		cli := proto.NewCLI(proto.TypeHandlePacket, 0, bytes[:bnum], destAddr)
		li.NodeChan <- cli
	}
}

func (li *LinkInterface) IsUp() bool {
	li.Mu.Lock()
	defer li.Mu.Unlock()
	return li.Status == "up"
}

// ****************************************************************************
// Send bytes through link
func (li *LinkInterface) SendPacket(packetBytes []byte) {
	// fmt.Printf("Link try to send a RIP to %v through port %v\n", li.MACRemote, li.MACRemote)
	// fmt.Printf("Link whose remote port is %v 's status is %v\n", li.MACRemote, li.Status)
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
