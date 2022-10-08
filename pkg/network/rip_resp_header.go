package network

import (
	"log"
	"net"
	"strconv"
	"strings"

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

func (node *Node) NewRIPRespHeader(IPLocal, IPRemote string, bodyLen int) *ipv4.Header {
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
		Src:      net.ParseIP(IPLocal),
		Dst:      net.ParseIP(IPRemote),
	}
	return header
}

func str2netIP(addr string) net.IP {
	res := make([]byte, 4)
	strs := strings.Split(addr, ".")
	for i, str := range strs {
		num, err := strconv.Atoi(str)
		if err != nil {
			log.Fatalln(err)
		}
		res[i] = byte(num)
	}
	// fmt.Println(res)
	return net.IP(res)
}

func netIP2str(IP net.IP) string {
	addr := IP.String()
	return addr
}
