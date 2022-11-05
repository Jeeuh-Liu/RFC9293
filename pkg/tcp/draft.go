package tcp

// import network

// //a port edge case : what if port is >= 65536

// // c ip port

// // Q1 can a socket be in SYN_SENT status forever?
// //   example c 192.168.0.2 999 when the remote listen on 1000
// //		why then open 999, the client's socket remains syn_sent
// //   example c 192.168.0.3 999 when the remote hasn't opened
// // port for VTCPConn should be atomic coutner
// func VConnect(addr net.IP, dstPort int16) (VTCPConn, error){
// 	//  first handshake
// 			//  payload := []byte{}
// 			//	srcPort := generate a port number
// 			//  seqNum := randomly generate a seq num x
// 			//  tcpHdr := (srcPort, dstPort, seqNum, header.TCPFlagSyn)
// 			//				default window size and acknum

// 		//  below is borrowed from demo
// 			//  checksum := iptcp_utils.ComputeTCPChecksum(&tcpHdr, sourceIp, destIp, payload)
// 			//  tcpHdr.Checksum = checksum
// 			// tcpHeaderBytes := make(header.TCP, iptcp_utils.TcpHeaderLen)
// 			// tcpHeaderBytes.Encode(&tcpHdr)
// 			// ipPacketPayload := make([]byte, 0, len(tcpHeaderBytes)+len(payload))
// 			// ipPacketPayload = append(ipPacketPayload, tcpHeaderBytes...)
// 			// ipPacketPayload = append(ipPacketPayload, []byte(payload)...)

// 		//  delegate to router to send
// 		//		if the dst is not reachable, return nil, "no routes to dst"
// 		//		*otherwise spawn a new socket, socket.State = SYN_SENT
// 		//		go socket.syn_sent()
// }
// //outside VConnect node.tuple2sockets[segment.remoteAddress::localPort] = socket

// func (*VTCPConn) syn_sent(){
// // 	for{
// // 		select{
// // 		case <- timeout:
// //			*resend
// // 		case <- receive_ack:
// //			if syn &  ack == x + 1
// //				-> update seq & state
// //				-> then return
// // 		}
// // 	}
// }
