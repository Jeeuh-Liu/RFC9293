package network

import (
	"log"

	"golang.org/x/net/ipv4"
)

type RIPReq struct {
	Header *ipv4.Header
	Body   *RIPReqBody
}

func (node *Node) NewRIPReq(IPLocal, IPRemote string) *RIPReq {
	rip := &RIPReq{}
	rip.Body = node.NewRIPReqBody(IPRemote)
	rip.Header = node.NewRIPReqHeader(IPLocal, IPRemote, len(rip.Body.Marshal()))
	return rip
}

func (rip *RIPReq) Marshal() []byte {
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

func UnmarshalRIPReq(bytes []byte) RIPReq {
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln(err)
	}
	body := UnmarshalReqBody(bytes[20:])
	rip := RIPReq{
		Header: header,
		Body:   body,
	}
	return rip
}
