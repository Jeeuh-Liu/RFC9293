package network

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"
	"tcpip/pkg/link"

	"github.com/google/netstack/tcpip/header"
	"golang.org/x/net/ipv4"
)

type RIPReq struct {
	Header *ipv4.Header
	Body   *RIPReqBody
}

func (node *Node) NewRIPReq(li *link.LinkInterface) *RIPReq {
	rip := &RIPReq{}
	rip.Body = node.NewRIPReqBody(li.IPRemote)
	rip.Header = node.NewRIPReqHeader(li.IPLocal, li.IPRemote, len(rip.Body.Marshal()))
	headerBytes, err := rip.Header.Marshal()
	if err != nil {
		log.Fatalln("Error marshalling header:  ", err)
	}
	rip.Header.Checksum = int(ComputeChecksum(headerBytes))
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

func ComputeChecksum(b []byte) uint16 {
	checksum := header.Checksum(b, 0)

	// Invert the checksum value.  Why is this necessary?
	// The checksum function in the library we're using seems
	// to have been built to plug into some other software that expects
	// to receive the complement of this value.
	// The reasons for this are unclear to me at the moment, but for now
	// take my word for it.  =)
	checksumInv := checksum ^ 0xffff

	return checksumInv
}

// ***********************************************************************
// Req Header
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
		Src:      net.ParseIP(IPLocal),
		Dst:      net.ParseIP(IPRemote),
	}
	return header
}

// ***********************************************************************
// Req Body
type RIPReqBody struct {
	// command + num_entries = 4 bytes
	Command     uint16
	Num_Entries uint16
	// one entry = 12 bytes
	Entries []Entry
}

func (node *Node) NewRIPReqBody(IPRemote string) *RIPReqBody {
	entries := []Entry{}
	body := &RIPReqBody{
		Command:     uint16(1),
		Num_Entries: uint16(0),
		Entries:     entries,
	}
	return body
}

func (body *RIPReqBody) Marshal() []byte {
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
	for _, entry := range body.Entries {
		bytes = append(bytes, entry.Marshal()...)
	}
	return bytes
}

func UnmarshalReqBody(bytes []byte) *RIPReqBody {
	command := uint16(binary.BigEndian.Uint16(bytes[:2]))
	num_entries := uint16(binary.BigEndian.Uint16(bytes[2:4]))
	entries := []Entry{}
	body := &RIPReqBody{
		Command:     command,
		Num_Entries: num_entries,
		Entries:     entries,
	}
	return body
}
