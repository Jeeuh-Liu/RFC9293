package tcp

import network


type VTCPListener Struct{
	State uint8
	acceptQueue chan *Segment
	upstreamChan chan *VTCPConn
	LocalPort uint16
}

type VTCPConn Struct{
	State uint8
	// buffer undecided
	expectedACK uint16
	LocalAddr uint32
	LocalPort uint16
	RemoteAddr uint32
	RemotePort uint16
}

//port2listeners map[port]*VTCPListener
//tuple2sockets map[remoteIP::remotePort::localPort]*VTCPConn

//a port
	// Spawn a socket, bind it to the given port -- call
	//	listener, err := VListen(port)
	//		port can already be bound -> print err
	//	on success
		//  port2listeners[port] = listener
		//  listener.State = LISTENER
		//  go listener.acceptLoop()

func (listener *VTCPListener) acceptLoop() error{
// 	for{
// 		segment := <- listener.acceptQueue
// 		conn, err := listener.accept()
// 		if err == nil{
// 			listener.upstreamChan <- conn
// 		}
// 	}
}

func (listener *VTCPListener) VAccept() (*VTCPConn, error){
	// TODO: just return a socket object?
}

func (node *Node) handleTCP(){
	// for{
	// 	select{
	// 		case segment := <- node.segmentChan:
	//			TODO: if the conn exists
	//			if socket, found := node.tuple2sockets[segment.SrcIP::SrcPort::DstPort]; found{
	//				socket.buffer <- segment
	// 			}
	// 			if listener, found := node.port2listeners[dst_port]; !found{
	// 				listener.acceptQueue <- segment		
	// 			}
	// 		case socket := <- node.newSocketChan:
	//			node.tuple2sockets[segment.SrcIP::SrcPort::DstPort] = socket
	// 			go conn.synRecv()
	// 	}
	// }
}

// c ip port

// Q1 can a socket be in SYN_SENT status forever?
//   example c 192.168.0.2 999 when the remote listen on 1000
//		why then open 999, the client's socket remains syn_sent
//   example c 192.168.0.3 999 when the remote hasn't opened
// port for VTCPConn should be atomic coutner
func VConnect(addr net.IP, dstPort int16) (VTCPConn, error){
	//  first handshake
			//  payload := []byte{}
			//	srcPort := generate a port number
			//  seqNum := randomly generate a seq num x
			//  tcpHdr := (srcPort, dstPort, seqNum, header.TCPFlagSyn)
			//				default window size and acknum

		//  below is borrowed from demo
			//  checksum := iptcp_utils.ComputeTCPChecksum(&tcpHdr, sourceIp, destIp, payload)
			//  tcpHdr.Checksum = checksum
			// tcpHeaderBytes := make(header.TCP, iptcp_utils.TcpHeaderLen)
			// tcpHeaderBytes.Encode(&tcpHdr)
			// ipPacketPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
			// ipPacketPayload = append(ipPacketPayload, tcpHeaderBytes...)
			// ipPacketPayload = append(ipPacketPayload, []byte(payload)...)
	
			
		//  delegate to router to send
		//		if the dst is not reachable, return nil, "no routes to dst"
		//		*otherwise spawn a new socket, socket.State = SYN_SENT
		//		go socket.syn_sent()
}
//outside VConnect node.tuple2sockets[segment.remoteAddress::localPort] = socket


func (*VTCPConn) syn_sent(){
// 	for{
// 		select{
// 		case <- timeout:
//			*resend	
// 		case <- receive_ack:
//			if syn &  ack == x + 1 
//				-> update seq & state
//				-> then return 
// 		}
// 	}
}

//Q: handle RST-?
func (*VTCPConn) syn_recv(){
	// 	for{
	// 		select{
	// 		case <- timeout:
	//			*resend	
	// 		case <- receive_ack:
	//			if ack == y + 1
	//				set state to established
	// 		}
	// 	}
}

func (*VTCPListener) VClose() error{
	//TODO: just remain empty? 
	//I know the node should remote the mapping of this listener
}