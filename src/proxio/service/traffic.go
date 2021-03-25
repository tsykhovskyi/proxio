package service

import (
	"proxio/traffic"
)

func TrafficStorage() *traffic.Storage {
	if services.trafficStorage == nil {
		services.trafficStorage = traffic.NewStorage()
	}

	return services.trafficStorage
}

func TrafficTracker() *traffic.TrafficTracker {
	if services.trafficTracker == nil {
		services.trafficTracker = traffic.NewTrafficTracker()
	}

	return services.trafficTracker
}
