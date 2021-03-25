package service

import (
	"proxio/repository"
	"proxio/traffic"
	"proxio/ws"
)

var services struct {
	session          *repository.SessionsRepo
	wsConnectionPool *ws.Pool
	trafficTracker   *traffic.TrafficTracker
	trafficStorage   *traffic.Storage
}
