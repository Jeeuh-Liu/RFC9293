package network

import (
	"bytes"
	"encoding/binary"
	"log"
)

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
	// fmt.Printf("Length of body [Command + Num_Entries] is %v bytes\n", len(bytes))
	for _, entry := range body.Entries {
		bytes = append(bytes, entry.Marshal()...)
		// fmt.Printf("Length of body [Command + Num_Entries] + entry is %v bytes\n", len(bytes))
	}
	// fmt.Printf("Length of body is %v bytes\n", len(bytes))
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
