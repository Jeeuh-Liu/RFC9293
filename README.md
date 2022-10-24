# How you abstract your link layer and its interfaces

The structure of Link Interface

| Variable name | Variable Type         | Functionality                         |
| ------------- | --------------------- | ------------------------------------- |
| Mu            | sync.Mutex            | protect Status of link interface      |
| ID            | uint8                 | id of                                 |
| MACLocal      | string                | MAC address of local node             |
| MACRemote     | string                | MAC address of remote node            |
| IPLocal       | string                | local IP address of the interface     |
| IPRemote      | string                | remote IP address of the interface    |
| Status        | string                | mark this link interface is up / down |
| LinkConn      | *net.UDPConn          | read / write bytes from/into the link |
| NodePktOpChan | chan *proto.NodePktOp | send bytes back to its node           |



# The thread model for your RIP implementation









# The steps you will need to process IP packets