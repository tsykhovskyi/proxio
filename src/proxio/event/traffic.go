package event

import "proxio/service"

func HandleTraffic() {
	go func() {
		for m := range service.TrafficTracker().GetTraffic() {
			service.TrafficStorage().Add(m)
			service.WsConnectionPool().BroadcastMessage(m.GetContext())
		}
	}()
}
