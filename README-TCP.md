# Design

- client interaction: use a goroutine to receive the cmd from the client, and delegate the cmd to each handle function
- Several components in tcp package: node(kernel), listener, normal_socket
  - listener contains a channel(SegRcvChan) to receive segment from the network layer, a channel to transfer pending connection to the kernel, a channel to cancel AcceptLoop when the listener is to be closed.
  - normal_socket consists of recv buffer,send buffer,(both of them run in a goroutine) a retransmit queue, three channels to communicate with the kernel(send out segment, recv segment, notify channel to delete itself from the socket table), and several condition variables, for example NonEmptyCond to indicate whehter the conn can read from a recv buffer. And a close channel to force state change(ESTAB->FIN_WAIT1, CLOSE_W -> LASTACK)
  - Kernel has a socket table to map tuples to conns or ports to listeners, and support opertaions like offer, delete, find
- Basically, we follow the TCP state machine to organize our code. Each state has a function. State transfer in our program is to jump to one function to the other, to simply the logic
- In terms of retransmission, the client will input all segments into retransmission queue and conn.seq will only be added when corresponding ackNum has been received by the client. In this way, we can utilize conn.seq to check whether current segment has been successfully acked by the server. If some packets loss happen, the client will retransmit this segment again.



# Performance Measurement

Reference node

|       | frame number | time  |
| ----- | ------------ | ----- |
| start | 41           | 24.01 |
| end   | 5518         | 24.65 |

Our node

|       | frame number | time  |
| ----- | ------------ | ----- |
| start | 24           | 12.12 |
| end   | 4479         | 12.41 |



# Packet Capture

The 3-way handshake

| frame number | packet sent by A (isn: 2596996162) | packet sent by B(isn: 55315)              |
| ------------ | ---------------------------------- | ----------------------------------------- |
| 62           | SYN, seq = 0, win = 65535          |                                           |
| 64           |                                    | SYN, ACK, seq = 0, ack = 1 window = 63335 |
| 66           | ACK, seq = 1, ack = 1, win = 65535 |                                           |



Segment sent and acknowledged

| frame number | packet sent by A (isn: 2596996162)     | packet sent by B(isn: 55315) |
| ------------ | -------------------------------------- | ---------------------------- |
| 834          | ACK, seq = 159343, ack = 1, len = 1285 |                              |
| 1268         |                                        | ACK, seq = 1, ack = 160628   |



One segment that is retransmitted

| frame number       | packet sent by A (isn: 2596996162)                 | packet sent by B(isn: 55315)          |
| ------------------ | -------------------------------------------------- | ------------------------------------- |
| 96                 | ACK, seq = 14136, ack = 1, win = 65535, len=1285   |                                       |
| 97(retransmission) | ACK, seq = 14136, ack = 1, win = 65535, len=1285   |                                       |
| 98                 | ACK, seq = 15421, ack = 1, win = 65535, len = 1285 |                                       |
| 125                |                                                    | ACK, seq = 1, ack= 16706, win = 50878 |



Connection Teardown

| frame number | packet sent by A (isn: 2596996162)             | packet sent by B(isn: 70841)                  |
| ------------ | ---------------------------------------------- | --------------------------------------------- |
| 10070        | FIN, ACK, seq = 1398105, ack = 1,  win = 65535 |                                               |
| 10072        |                                                | ACK, seq = 1, ack = 1398106, win = 65535      |
| 10073        |                                                | FIN, ACK, seq = 1, ack = 1398106, win = 65535 |
| 10076        | ACK, seq = 1398106, ack = 2, win=65535         |                                               |

