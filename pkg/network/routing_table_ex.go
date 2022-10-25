package network

import "time"

func (rt *RoutingTable) CheckRouteEx(destIP string) {
	if time.Now().After(rt.RemoteDest2ExTime[destIP]) {
		// fmt.Println(destIP, "Del 275")
		newRoute := rt.DestIP2Route[destIP]
		newRoute.Cost = 16
		rt.UpdateRoutesAndBroadcastTU(newRoute, destIP)
	}
}

// Update Expiration time of routes
func (rt *RoutingTable) UpdateExTime(destIP string) {
	rt.RemoteDest2ExTime[destIP] = time.Now().Add(12 * time.Second)
	go rt.SendExTimeCLI(destIP)
}
