package network

import (
	"golang.org/x/net/ipv4"
)

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

func (node *Node) NewRIPReqHeader(IPLocal, IPRemote string, bodyLen int) *ipv4.Header {
	header := &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + bodyLen,
		Flags:    0,
		FragOff:  0,
		TTL:      16,
		Protocol: 200,
		Checksum: 0,
		Src:      str2netIP(IPLocal),
		Dst:      str2netIP(IPRemote),
	}
	return header
}
