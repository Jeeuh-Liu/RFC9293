package network

import (
	"fmt"
	"strconv"
	"tcpip/pkg/proto"
	"tcpip/pkg/tcp"

	"github.com/google/netstack/tcpip/header"
)

func (node *Node) handleTCP() {
	for {
		segment := <-node.segRecvChan
		tuple := segment.FormTuple()
		if conn := node.socketTable.FindConn(tuple); conn != nil {
			conn.Buffer <- segment
		}
		dstPort := segment.TCPhdr.DstPort
		listener := node.socketTable.FindListener(dstPort)
		if listener != nil {
			listener.AcceptQueue <- segment
		}
	}
}

func (node *Node) handleCreateListener(msg *proto.NodeCLI) {
	val, _ := strconv.Atoi(msg.Msg)
	port := uint16(val)
	if node.socketTable.FindListener(port) != nil {
		fmt.Printf("Cannot assign requested address\n")
	} else {
		listener := node.socketTable.OfferListener(port)
		go node.acceptConn(listener)
	}
}

// a port -> listener  -> go node.acceptConn(listener)
func (node *Node) acceptConn(listener *tcp.VTCPListener) {
	for {
		conn, err := listener.VAccept()
		if err != nil {
			continue
		}
		tuple := conn.GetTuple()
		conn.Upstream = node.segSendChan
		node.socketTable.OfferConn(tuple, conn)
		//syn : 1
		go conn.SynRecv()
	}
}

func (node *Node) sendOutSegment(msg *proto.SegmentMsg) {
	conn := node.socketTable.FindConnByID(msg.SocketID)
	hdr := msg.Seg.TCPhdr
	payload := msg.Seg.Payload
	checksum := proto.ComputeTCPChecksum(hdr, conn.LocalAddr, conn.RemoteAddr, payload)
	hdr.Checksum = checksum
	tcpHeaderBytes := make(header.TCP, proto.TcpHeaderLen)
	tcpHeaderBytes.Encode(hdr)
	iPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
	iPayload = append(iPayload, tcpHeaderBytes...)
	iPayload = append(iPayload, []byte(payload)...)
	proto.PrintHex(iPayload)
	node.RT.SendPacket(conn.RemoteAddr.String(), proto.PROTOCOL_TCP, string(iPayload))
}
