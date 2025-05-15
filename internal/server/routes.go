package server

import (
	"load-balancer/internal/server/middleware"
	"net/http"

	"load-balancer/internal/service"
)

// RegisterRoutes — регистрирует HTTP-роуты: прокси с рейт-лимитом, а также CRUD-эндпоинты для клиентов.
func RegisterRoutes(proxySvc *service.ProxyService, clientSvc *service.ClientService, middleware *middleware.RateLimitMiddleware) {
	http.HandleFunc("/", middleware.Middleware(proxySvc.ProxyHandler()))
	http.HandleFunc("POST /clients", clientSvc.CreateClientHandler())
	http.HandleFunc("GET /clients/{id}", clientSvc.GetClientHandler())
	http.HandleFunc("PATCH /clients/{id}", clientSvc.UpdateClientHandler())
	http.HandleFunc("DELETE /clients/{id}", clientSvc.DeleteClientHandler())
}
