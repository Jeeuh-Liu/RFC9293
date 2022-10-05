package ip

const (
	LI       = uint8(0)
	SetUpT   = uint8(1)
	SetDownT = uint8(2)
	Quit     = uint8(3)
)

type CLI struct {
	CLIType uint8
	ID      uint8
}

func NewCLI(cliType, id uint8) *CLI {
	cli := &CLI{
		CLIType: cliType,
		ID:      id,
	}
	return cli
}
