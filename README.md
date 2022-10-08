# Command



| Command                          | Description                                                  |
| :------------------------------- | :----------------------------------------------------------- |
| `interfaces`, `li`               | Prints information about each interface, one per line.       |
| `interfaces <file>`, `li <file>` | Print information about each interface, one per line, to the destination file. Overrides the file if it exists. |
| `routes`, `lr`                   | Print information about the route to each known destination, one per line. |
| `routes <file>`, `lr <file>`     | Print information about the route to each known destination, one per line, to the destination file. Overwrites the file if it exists. |
| `down <integer>`                 | Bring an interface with ID `<integer>` “down”.               |
| `up <integer>`                   | Bring an interface with ID `<integer>` “up” (it must be an existing interface, probably one you brought down). |
| `send <vip> <proto> <string>`    | Send an IP packet with protocol `<proto>` (an integer) to the virtual IP address `<vip>` (dotted quad notation). The payload is simply the characters of `<string>` (as in Snowcast, do not null-terminate this). |
| `q`                              | Quit the node by cleaning up used resources.                 |





# Test

start node

```shell
reference node + inx

./tools/ref_node ./nets/routeAggregation/tree/A.lnx
./tools/ref_node ./nets/routeAggregation/tree/B.lnx
./tools/ref_node ./nets/routeAggregation/tree/C.lnx
```

print all link interfaces

```shell
li
```

test down/up

```shell
up 0
li
down 0
li
up 0
```



print all routers

```shell
# my nodeA
# my nodeB
# ref nodeC

lr
```



send packets

```shell
send <vip> <proto> <string>    

#A send to B
send 10.0.0.14 0 "Hello from A"

#A send to C
send 10.0.0.10 0 "Hello from A"
```



# RIP

## RIP struct

### Header

| element  |                       | functionality                                     |
| -------- | --------------------- | ------------------------------------------------- |
| Protocol | 200                   |                                                   |
| Len      | 120                   | avoid err "Header Marshal Error header too short" |
| Src      | IP of local interface | next hop for new route                            |



### Body

| Element       | Type    | value                                                        |
| ------------- | ------- | ------------------------------------------------------------ |
| command       | uint16  | `1` for a request of routing information, and `2` for a response |
| num_entries   | uint16  | len(entries)                                                 |
| entries       | []Entry | all valid entries that do not go back to source of IP        |
| Entry.cost    | uint32  | current route entry + 1                                      |
| Entry.address | uint32  | Dest of current route entry                                  |
| Entry.mask    | uint32  | 1 << 32 - 1                                                  |



## min_cost of a routing entry

| metadata          | type              |                                     |
| ----------------- | ----------------- | ----------------------------------- |
| RemoteDestIP2Cost | map[string]uint32 | record min cost of remote dest addr |

When we receive an RIP msg

we check all of its route entries

new cost = entry.cost + 1

if dest addr exits in RemoteDestIP2Cost && newCost >= oldCost: ignore it

else: create a newroute and store it



## Split Horizon with Poisoned Reverse

| metadata           | type              |                                 |
| ------------------ | ----------------- | ------------------------------- |
| RemoteDestIP2SrcIP | map[string]string | record src ip of remote dest ip |

When we are sending out RIP

if next ip of this dest ip is src ip, ignore this entry 

else put the entry into body of the packet



## Expiration of a routing entry





## Triggered updates
