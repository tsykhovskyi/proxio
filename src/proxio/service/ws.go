package service

import (
	"proxio/ws"
)

func WsConnectionPool() *ws.Pool {
	if services.wsConnectionPool == nil {
		services.wsConnectionPool = ws.NewConnectionPool()
	}

	return services.wsConnectionPool
}
