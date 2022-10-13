package proto

// API of node
const (
	// Command Line Interface
	LI       = uint8(0)
	SetUpT   = uint8(1)
	SetDownT = uint8(2)
	Quit     = uint8(3)
	LR       = uint8(4)
	// network pass Packet to link
	TypeBroadcastRIPReq  = uint8(5)
	TypeBroadcastRIPResp = uint8(6)
	// Remote Route Expiration
	TypeRouteEx = uint8(8)
	// Send Packet to Link
	TypeSendPacket = uint8(9)
	// Link pass packet back to network
	TypeReceivePacket = uint8(7)
)

// Broadcast RIP Request and RIP Response
type NodeBC struct {
	OpType uint8
	ID     uint8
	// packet: bytes of body
	Bytes []byte
	// dest IP (which can be used to send packet)
	DestIP string
	// Send Packet
	ProtoID int
	Msg     string
}

func NewNodeBC(opType, id uint8, bytes []byte, destIP string, protoID int, msg string) *NodeBC {
	nodeBC := &NodeBC{
		OpType:  opType,
		ID:      id,
		Bytes:   bytes,
		DestIP:  destIP,
		ProtoID: protoID,
		Msg:     msg,
	}
	return nodeBC
}
