# How you abstract your link layer and its interfaces

Each link interface includes the vitual link (UDP socket) to send/receive bytes to/from neighbors and metadata about IP address and MAC address of linked nodes.



The structure of Link Interface looks like this:

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
| NodePktOpChan | chan *proto.NodePktOp | send bytes of packet back to its node |



# The thread model for your RIP implementation

## Broadcast RIP packets

- When the node comes online, broadcast RIP request packets

- After coming online, broadcast RIP response packets every 5 seconds periodically 

  - If the route entry is sent back to source node, set the cost of route entry to 16(infinity)

- If 1 route entry gets updated, broadcast triggered updates to neighbors

- If 1 route entry expires, broadcast triggered udpate( cost of route entry == 16) to neighbors

- If status of 1 link is set to "down", broadcast triggered udpate( cost of route entry == 16) to neighbors

  

## Handle RIP packet

- Validity
  - Check whether the checksum in header is valid
  - Check if TTL == 0
- RIP Packet
  - if destIP is local IP of current node, it will not expire
  - if destIP does not exist in routing table of current node, add the route to routing table and reset its expiration time
  - if newCost < oldCost, update cost of this route to smaller cost and reset its expiration time
  - if newCost > oldCost and newNextIPAddress == oldNextIPAddress, update cost of this route to larger one and reset its expiration time. If newCost == 16, delete newly added route from routing table because this is a triggered update caused by expiration of the route
  - if newCost > oldCost and newNextIPAddress != oldNextIPAddress, ignore this route entry
  - If newCost == oldCost, reset the expiration time



# The steps you will need to process IP packets

## Handle IP Packet

- Validity
  - Check whether the checksum in header is valid
  - Check if TTL == 0
- Test Packet
  - If destIP is one of local IP addresses of current node, print out this test msg
  - If destIP matches any route in the routing table, send it though the corresponding link interface
  - If destIP does not any route, stop routing