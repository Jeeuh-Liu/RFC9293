package proto

import (
	"log"
	"net"

	"golang.org/x/net/ipv4"
)

type Test struct {
	Header *ipv4.Header
	Body   []byte
}

func NewTest(IPSrc, IPDest, msg string) *Test {
	body := []byte(msg)
	header := NewTestHeader(IPSrc, IPDest, len(body))
	test := &Test{
		Header: header,
		Body:   body,
	}
	return test
}

func (test *Test) Marshal() []byte {
	bytes, err := test.Header.Marshal()
	if err != nil {
		log.Fatalln("Header Marshal Error", err)
	}
	bytes = append(bytes, test.Body...)
	// fmt.Printf("Total length of rip is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalTest(bytes []byte) *Test {
	header, err := ipv4.ParseHeader(bytes[:20])
	if err != nil {
		log.Fatalln(err)
	}
	body := bytes[20:]

	test := &Test{
		Header: header,
		Body:   body,
	}
	return test
}

// *******************************************************************
// Test Header
func NewTestHeader(IPSrc, IPDest string, bodyLen int) *ipv4.Header {
	return &ipv4.Header{
		Version:  4,
		Len:      20,
		TOS:      0,
		TotalLen: 20 + bodyLen,
		Flags:    0,
		FragOff:  0,
		TTL:      16,
		Protocol: 0,
		Checksum: 0,
		Src:      net.ParseIP(IPSrc),
		Dst:      net.ParseIP(IPDest),
	}
}
