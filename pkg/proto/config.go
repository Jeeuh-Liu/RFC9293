package proto

// API of node
const (
	// Command Line Interface
	CLI_SETUP   = uint8(0)
	CLI_SETDOWN = uint8(1)
	CLI_QUIT    = uint8(2)
	CLI_LI      = uint8(3)
	CLI_LR      = uint8(4)
	CLI_LIFILE  = uint8(5)
	CLI_LRFILE  = uint8(6)
	// network pass Packet to link
	MESSAGE_BCRIPREQ  = uint8(7)
	MESSAGE_BCRIPRESP = uint8(8)
	// Remote Route Expiration
	MESSAGE_ROUTEEX = uint8(9)
	// Send Packet to Link
	MESSAGE_SENDPKT = uint8(10)
	// Link pass packet back to network
	MESSAGE_REVPKT = uint8(11)
)
