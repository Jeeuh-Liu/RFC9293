package proto

// API of node
const (
	// Command Line Interface
	SetUpT   = uint8(0)
	SetDownT = uint8(1)
	Quit     = uint8(2)
	LI       = uint8(3)
	LR       = uint8(4)
	LIFILE   = uint8(5)
	LRFILE   = uint8(6)
	// network pass Packet to link
	TypeBroadcastRIPReq  = uint8(7)
	TypeBroadcastRIPResp = uint8(9)
	// Remote Route Expiration
	TypeRouteEx = uint8(9)
	// Send Packet to Link
	TypeSendPacket = uint8(10)
	// Link pass packet back to network
	TypeReceivePacket = uint8(11)
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
