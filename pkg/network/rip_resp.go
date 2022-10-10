package network

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"tcpip/pkg/link"

	"golang.org/x/net/ipv4"
)

type RIPResp struct {
	Header *ipv4.Header
	Body   *RIPRespBody
}

func (node *Node) NewRIPResp(li *link.LinkInterface) *RIPResp {
	rip := &RIPResp{}
	rip.Body = node.NewRIPRespBody(li.IPRemote)
	rip.Header = node.NewRIPRespHeader(li.IPLocal, li.IPRemote, len(rip.Body.Marshal()))
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

// ************************************************************************
// Header
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

// ************************************************************************
// Body

type RIPRespBody struct {
	// command + num_entries = 4 bytes
	Command     uint16
	Num_Entries uint16
	// one entry = 12 bytes
	Entries []Entry
}

func (node *Node) NewRIPRespBody(IPRemote string) *RIPRespBody {
	entries := []Entry{}
	for _, route := range node.DestIP2Route {
		// if route.next == src of route.dest -> ignore this route entry
		entry := NewEntry(route)
		if srcIP, ok := node.RemoteDestIP2SrcIP[route.Dest]; ok && srcIP == IPRemote {
			entry.Cost = 16
		}
		entries = append(entries, entry)
		// fmt.Println(entries)
	}
	body := &RIPRespBody{
		Command:     uint16(2),
		Num_Entries: uint16(len(entries)),
		Entries:     entries,
	}
	return body
}

func (body *RIPRespBody) Marshal() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, body.Command)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, body.Num_Entries)
	if err != nil {
		log.Fatalln(err)
	}
	bytes := buf.Bytes()
	// fmt.Printf("Length of body [Command + Num_Entries] is %v bytes\n", len(bytes))
	for _, entry := range body.Entries {
		bytes = append(bytes, entry.Marshal()...)
		// fmt.Printf("Length of body [Command + Num_Entries] + entry is %v bytes\n", len(bytes))
	}
	// fmt.Printf("Length of body is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalRespBody(bytes []byte) *RIPRespBody {
	command := uint16(binary.BigEndian.Uint16(bytes[:2]))
	num_entries := uint16(binary.BigEndian.Uint16(bytes[2:4]))
	entries := []Entry{}
	for i := 0; i < int(num_entries); i++ {
		start, end := 4+i*12, 4+(i+1)*12
		entry := UnmarshalEntry(bytes[start:end])
		entries = append(entries, entry)
	}
	body := &RIPRespBody{
		Command:     command,
		Num_Entries: num_entries,
		Entries:     entries,
	}
	return body
}
