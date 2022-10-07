package network

// API of node
const (
	// command from user
	LI       = uint8(0)
	SetUpT   = uint8(1)
	SetDownT = uint8(2)
	Quit     = uint8(3)
	LR       = uint8(4)
	// packet
	RIP  = uint8(5)
	TEST = uint8(6)
)

type CLI struct {
	CLIType uint8
	ID      uint8
	// packet: bytes of body
	Bytes []byte
}

func NewCLI(cliType, id uint8, bytes []byte) *CLI {
	cli := &CLI{
		CLIType: cliType,
		ID:      id,
		Bytes:   bytes,
	}
	return cli
}
