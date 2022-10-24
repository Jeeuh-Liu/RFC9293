package proto

type NodeCLI struct {
	CLIType uint8
	ID      uint8
	// packet: bytes of body
	Bytes []byte
	// dest IP (which can be used to send packet)
	DestIP string
	// Send Packet
	ProtoID  int
	Msg      string
	Filename string
}

func NewNodeCLI(cliType, id uint8, bytes []byte, destIP string, protoID int, msg string, filename string) *NodeCLI {
	nodeCLI := &NodeCLI{
		CLIType:  cliType,
		ID:       id,
		Bytes:    bytes,
		DestIP:   destIP,
		ProtoID:  protoID,
		Msg:      msg,
		Filename: filename,
	}
	return nodeCLI
}
