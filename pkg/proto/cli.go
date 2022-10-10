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
	// Link pass packet back to network
	TypeHandlePacket = uint8(7)
	// Handle RIP Resp
	TypeHandleRIPResp = uint8(8)
	// Expiration
	TypeRouteEx = uint8(9)
)

type CLI struct {
	CLIType uint8
	ID      uint8
	// packet: bytes of body
	Bytes []byte
	// dest IP
	DestIP string
}

func NewCLI(cliType, id uint8, bytes []byte, destIP string) *CLI {
	cli := &CLI{
		CLIType: cliType,
		ID:      id,
		Bytes:   bytes,
		DestIP:  destIP,
	}
	return cli
}
