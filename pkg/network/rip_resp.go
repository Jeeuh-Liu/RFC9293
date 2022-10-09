package network

import (
	"log"

	"golang.org/x/net/ipv4"
)

type RIPResp struct {
	Header *ipv4.Header
	Body   *RIPRespBody
}

func (node *Node) NewRIPResp(IPLocal, IPRemote string) *RIPResp {
	rip := &RIPResp{}
	rip.Body = node.NewRIPRespBody(IPRemote)
	rip.Header = node.NewRIPRespHeader(IPLocal, IPRemote, len(rip.Body.Marshal()))
	headerBytes, err := rip.Header.Marshal()
	if err != nil {
		log.Fatalln("Error marshalling header:  ", err)
	}
	rip.Header.Checksum = int(ComputeChecksum(headerBytes))
	return rip
}

func (rip *RIPResp) Marshal() []byte {
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

func UnmarshalRIPResp(bytes []byte) RIPResp {
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln(err)
	}
	body := UnmarshalRespBody(bytes[20:])
	rip := RIPResp{
		Header: header,
		Body:   body,
	}
	return rip
}
