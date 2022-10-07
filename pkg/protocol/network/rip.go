package network

import (
	"fmt"
	"log"

	"golang.org/x/net/ipv4"
)

type RIPPacket struct {
	Header *ipv4.Header
	Body   *RIPBody
}

func (node *Node) NewRIP() *RIPPacket {
	rip := &RIPPacket{}
	rip.Header = node.NewRIPHeader()
	rip.Body = node.NewRIPBody()
	return rip
}

func (rip *RIPPacket) Marshal() []byte {
	bytes, err := rip.Header.Marshal()
	// num of bytes in header is 20 bytes
	// fmt.Printf("num of bytes of Header is %v\n", len(bytes))
	if err != nil {
		log.Fatalln("Header Marshal Error", err)
	}
	bytes = append(bytes, rip.Body.Marshal()...)
	fmt.Printf("Total length of rip is %v bytes\n", len(bytes))
	return bytes
}

// ************************************************************************
// Header

/*
	type Header struct {
	Version  int         // protocol version
	Len      int         // header length
	TOS      int         // type-of-service
	TotalLen int         // packet total length
	ID       int         // identification
	Flags    HeaderFlags // flags
	FragOff  int         // fragment offset
	TTL      int         // time-to-live
	Protocol int         // next protocol
	Checksum int         // checksum
	Src      net.IP      // source address
	Dst      net.IP      // destination address
	Options  []byte      // options, extension headers
}
*/

func (node *Node) NewRIPHeader() *ipv4.Header {
	header := &ipv4.Header{
		Version:  0,
		Len:      128,
		TOS:      0,
		TotalLen: 0,
		Flags:    0,
		FragOff:  0,
		TTL:      0,
		Protocol: 200,
		Checksum: 0,
		Src:      make([]byte, 4),
		Dst:      make([]byte, 4),
	}
	return header
}
