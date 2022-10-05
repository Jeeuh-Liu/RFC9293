# Demand



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

```

send packets

```shell
send <vip> <proto> <string>    

#A send to B
send 10.0.0.14 0 "Hello from A"

#A send to C
send 10.0.0.10 0 "Hello from A"
```

