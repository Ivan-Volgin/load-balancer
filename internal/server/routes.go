package server

import (
	"net/http"

	"load-balancer/internal/service"
)

func RegisterRoutes(proxySvc *service.ProxyService, clientSvc *service.ClientService) {
	http.HandleFunc("/", proxySvc.ProxyHandler())
	http.HandleFunc("POST /clients", clientSvc.CreateClientHandler())
	http.HandleFunc("GET /clients/{id}", clientSvc.GetClientHandler())
	http.HandleFunc("PATCH /clients/{id}", clientSvc.UpdateClientHandler())
	http.HandleFunc("DELETE /clients/{id}", clientSvc.DeleteClientHandler())
}
