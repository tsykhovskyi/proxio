package service

import "proxio/repository"

func Session() *repository.SessionsRepo {
	if services.session == nil {
		services.session = repository.NewSessionsRepo()
		services.session.Populate()
	}

	return services.session
}
