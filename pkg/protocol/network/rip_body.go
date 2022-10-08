package network

import (
	"bytes"
	"encoding/binary"
	"log"
)

type RIPBody struct {
	// command + num_entries = 4 bytes
	command     uint16
	num_entries uint16
	// one entry = 12 bytes
	entries []Entry
}

func (node *Node) NewRIPBody(IPRemote string) *RIPBody {
	entries := []Entry{}
	for _, route := range node.Routes {
		// if route.next == src of route.dest -> ignore this route entry
		if srcIP, ok := node.RemoteDestIP2SrcIP[route.Dest]; ok && srcIP == IPRemote {
			continue
		}
		entry := NewEntry(route)
		entries = append(entries, entry)
		// fmt.Println(entries)
	}
	num_entries := len(entries)
	body := &RIPBody{
		command:     1,
		num_entries: uint16(num_entries),
		entries:     entries,
	}
	return body
}

func (body *RIPBody) Marshal() []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, body.command)
	if err != nil {
		log.Fatalln(err)
	}
	err = binary.Write(buf, binary.BigEndian, body.num_entries)
	if err != nil {
		log.Fatalln(err)
	}
	bytes := buf.Bytes()
	// fmt.Printf("Length of body [Command + Num_Entries] is %v bytes\n", len(bytes))
	for _, entry := range body.entries {
		bytes = append(bytes, entry.Marshal()...)
		// fmt.Printf("Length of body [Command + Num_Entries] + entry is %v bytes\n", len(bytes))
	}
	// fmt.Printf("Length of body is %v bytes\n", len(bytes))
	return bytes
}

func UnmarshalBody(bytes []byte) *RIPBody {
	command := uint16(binary.BigEndian.Uint16(bytes[:2]))
	num_entries := uint16(binary.BigEndian.Uint16(bytes[2:4]))
	entries := []Entry{}
	for i := 0; i < int(num_entries); i++ {
		start, end := 4+i*12, 4+(i+1)*12
		entry := UnmarshalEntry(bytes[start:end])
		entries = append(entries, entry)
	}
	body := &RIPBody{
		command:     command,
		num_entries: num_entries,
		entries:     entries,
	}
	return body
}
