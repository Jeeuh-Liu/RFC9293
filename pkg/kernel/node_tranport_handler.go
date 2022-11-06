package kernel

import (
	"fmt"
	"net"
	"strconv"
	"tcpip/pkg/myDebug"
	"tcpip/pkg/proto"
	"tcpip/pkg/tcp"

	"github.com/google/netstack/tcpip/header"
)

func (node *Node) handleTCPSegment() {
	for {
		segment := <-node.segRecvChan
		tuple := segment.FormTuple()
		// fmt.Println(tuple)
		myDebug.Debugln("Receive a TCP packet from %v", tuple)
		// 2nd & 3rd handshake
		if conn := node.socketTable.FindConn(tuple); conn != nil {
			myDebug.Debugln("%v:%v receive packet from %v:%v, SEQ: %v, ACK %v",
				conn.LocalAddr.String(), conn.LocalPort, conn.RemoteAddr.String(),
				conn.RemotePort, segment.TCPhdr.SeqNum, segment.TCPhdr.AckNum)

			conn.SegRcvChan <- segment
			continue
		}
		// 1st handshake
		dstPort := segment.TCPhdr.DstPort
		listener := node.socketTable.FindListener(dstPort)
		if listener != nil {
			listener.SegRcvChan <- segment
		}
	}
}

// *****************************************************************************************
// Handle Create Listener
func (node *Node) handleCreateListener(msg *proto.NodeCLI) {
	val, _ := strconv.Atoi(msg.Msg)
	port := uint16(val)
	if node.socketTable.FindListener(port) != nil {
		fmt.Printf("Cannot assign requested address\n")
	} else {
		listener := node.socketTable.OfferListener(port)
		go node.NodeAcceptLoop(listener)
	}
}

// a port -> listener  -> go node.acceptConn(listener)
func (node *Node) NodeAcceptLoop(listener *tcp.VTCPListener) {
	for {
		conn, err := listener.VAccept()
		if err != nil {
			continue
		}
		conn.NodeSegSendChan = node.segSendChan
		node.socketTable.OfferConn(conn)
		//syn : 1
		conn.VTCPConnSynRev()
	}
}

// *****************************************************************************************
// Handle Create Listener
func (node *Node) HandleCreateConn(nodeCLI *proto.NodeCLI) {
	// Create a Normal Socket
	srcIP := node.RT.FindSrcIPAddr(nodeCLI.DestIP)
	if srcIP == "no" {
		fmt.Println("v_connect() error: No route to host")
		return
	}
	conn := tcp.NewNormalSocket(0, nodeCLI.DestPort, node.socketTable.ConnPort, net.ParseIP(nodeCLI.DestIP), net.ParseIP(srcIP))
	conn.NodeSegSendChan = node.segSendChan
	node.socketTable.OfferConn(conn)
	fmt.Println("New Socket has been created")
	// Send SYN
	conn.VTCPConnSynSend()
}

// *****************************************************************************************
func (node *Node) HandleSendOutSegment(seg *proto.Segment) {
	hdr := seg.TCPhdr
	payload := seg.Payload
	tcpHeaderBytes := make(header.TCP, proto.TcpHeaderLen)
	tcpHeaderBytes.Encode(hdr)
	iPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
	iPayload = append(iPayload, tcpHeaderBytes...)
	iPayload = append(iPayload, []byte(payload)...)
	// proto.PrintHex(iPayload)
	node.RT.SendTCPPacket(seg.IPhdr.Src.String(), seg.IPhdr.Dst.String(), string(iPayload))
}
