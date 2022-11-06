package kernel

import (
	"fmt"
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
		tuple := conn.GetTuple()
		conn.NodeSegSendChan = node.segSendChan
		node.socketTable.OfferConn(tuple, conn)
		//syn : 1
		conn.VTCPConnSynHandler()
	}
}

// *****************************************************************************************
func (node *Node) sendOutSegment(seg *proto.Segment) {
	hdr := seg.TCPhdr
	payload := seg.Payload
	checksum := proto.ComputeTCPChecksum(hdr, seg.IPhdr.Src, seg.IPhdr.Dst, payload)
	hdr.Checksum = checksum
	tcpHeaderBytes := make(header.TCP, proto.TcpHeaderLen)
	tcpHeaderBytes.Encode(hdr)
	iPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
	iPayload = append(iPayload, tcpHeaderBytes...)
	iPayload = append(iPayload, []byte(payload)...)
	// proto.PrintHex(iPayload)
	node.RT.SendTCPPacket(seg.IPhdr.Dst.String(), proto.PROTOCOL_TCP, string(iPayload))
}
