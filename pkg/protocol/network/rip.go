package network

import (
	"log"

	"golang.org/x/net/ipv4"
)

type RIPPacket struct {
	Header *ipv4.Header
	Body   *RIPBody
}

func (node *Node) NewRIP(IPLocal, IPRemote string) *RIPPacket {
	rip := &RIPPacket{}
	rip.Header = node.NewRIPHeader(IPLocal)
	rip.Body = node.NewRIPBody(IPRemote)
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
	// fmt.Printf("Total length of rip is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalRIP(bytes []byte) RIPPacket {
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln(err)
	}
	body := UnmarshalBody(bytes[20:])
	rip := RIPPacket{
		Header: header,
		Body:   body,
	}
	return rip
}
