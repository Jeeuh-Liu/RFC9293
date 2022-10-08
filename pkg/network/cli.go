package network

// API of node
const (
	// Command from user
	LI       = uint8(0)
	SetUpT   = uint8(1)
	SetDownT = uint8(2)
	Quit     = uint8(3)
	LR       = uint8(4)
	// Packet
	RIPReqBroadcast  = uint8(5)
	RIPRespBroadcast = uint8(6)
	RIPReqHandle     = uint8(7)
	RIPRespHandle    = uint8(8)
	// Expiration
	RouteEx = uint8(9)
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
