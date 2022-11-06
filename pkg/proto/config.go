package proto

// API of node
const (
	// Command Line Interface
	CLI_SETUP          = uint8(0)
	CLI_SETDOWN        = uint8(1)
	CLI_QUIT           = uint8(2)
	CLI_LI             = uint8(3)
	CLI_LIFILE         = uint8(5)
	CLI_LR             = uint8(4)
	CLI_LRFILE         = uint8(6)
	CLI_LS             = uint8(7)
	CLI_LSFILE         = uint8(8)
	CLI_CREATELISTENER = uint8(9)
	CLI_CREATECONN     = uint8(10)
	CLI_SENDSEGMENT    = uint8(11)
	CLI_RECVSEGMENT    = uint8(12)
	// network pass Packet to link
	MESSAGE_BCRIPREQ  = uint8(20)
	MESSAGE_BCRIPRESP = uint8(21)
	// Remote Route Expiration
	MESSAGE_ROUTEEX = uint8(22)
	// Send Packet to Link
	MESSAGE_SENDPKT = uint8(23)
	// Link pass packet back to network
	MESSAGE_REVPKT = uint8(24)

	PROTOCOL_RIP        = 200
	PROTOCOL_TESTPACKET = 0
	PROTOCOL_TCP        = 6

	LISTENER  = "LISTENER"
	SYN_RECV  = "SYN_RECV"
	ESTABLISH = "ESTABLISH"

	// the first port we allocate for conn
	FIRST_PORT = 0
)
