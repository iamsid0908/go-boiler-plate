package route

import (
	"core/config"
	"core/domain"
	"core/handler"
	"core/service"
)

type AppModel struct {
	Health handler.HealthHandler
	User   handler.UserHandler
	Auth   handler.AuthHandler
}

func App() AppModel {
	// Initialize queue client
	cfg := config.GetConfig()
	redisAddr := cfg.RedisAddr
	if redisAddr == "" {
		redisAddr = "localhost:6379" // default
	}

	//domain
	healthDomain := &domain.HealthDomainCtx{}
	authDomain := &domain.AuthDomainCtx{}
	userDomain := &domain.UserDomainCtx{}

	//service
	healthService := service.HealthService{
		HealthDomain: healthDomain,
	}
	userService := service.UserService{
		UserDomain: userDomain,
	}
	authService := service.AuthService{
		AuthDomain: authDomain,
		UserDomain: userDomain,
	}

	//handler
	healthHandler := handler.HealthHandler{
		HealthService: healthService,
	}
	userHandler := handler.UserHandler{
		UserService: userService,
	}
	authHandler := handler.AuthHandler{
		AuthService: authService,
	}

	return AppModel{
		Health: healthHandler,
		User:   userHandler,
		Auth:   authHandler,
	}
}
