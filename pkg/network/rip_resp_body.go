package network

import (
	"bytes"
	"encoding/binary"
	"log"
)

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
