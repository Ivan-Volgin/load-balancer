package service

import (
	"encoding/json"
	"load-balancer/internal/models"
	"load-balancer/internal/repo"
	"net/http"
	"time"
)

type ClientService struct {
	repo repo.Repository
}

func NewClientService(repo repo.Repository) *ClientService {
	return &ClientService{repo: repo}
}

func (cs *ClientService) CreateClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RateLimitClient
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.ClientID == "" || req.Capacity <= 0 || req.RatePerSecond <= 0 {
			writeJSONError(w, http.StatusBadRequest, "invalid client data")
		}

		req.Tokens = req.Capacity
		req.LastRefillAt = time.Now().Unix()

		if err := cs.repo.CreateClient(r.Context(), req); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to create client")
		}

		writeJSONResponse(w, http.StatusCreated, map[string]string{"status": "created", "clientID": req.ClientID})
	}
}

func (cs *ClientService) GetClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		client, err := cs.repo.GetClientByID(r.Context(), id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to get client")
			return
		}
		if client == nil {
			writeJSONError(w, http.StatusNotFound, "client not found")
			return
		}
		writeJSONResponse(w, http.StatusOK, client)
	}
}

func (cs *ClientService) UpdateClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if len(updates) == 0 {
			writeJSONError(w, http.StatusBadRequest, "no fields to update")
			return
		}

		existingClient, err := cs.repo.GetClientByID(r.Context(), id)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to get client")
			return
		}
		if existingClient == nil {
			writeJSONError(w, http.StatusNotFound, "client not found")
			return
		}

		if cap, ok := updates["capacity"].(float64); ok {
			existingClient.Capacity = int64(cap)
		}
		if rate, ok := updates["rate_per_second"].(float64); ok {
			existingClient.RatePerSecond = int64(rate)
		}
		if tokens, ok := updates["tokens"].(float64); ok {
			existingClient.Tokens = int64(tokens)
			existingClient.LastRefillAt = time.Now().Unix()
		}

		if err := cs.repo.UpdateClient(r.Context(), *existingClient); err != nil {
			writeJSONResponse(w, http.StatusInternalServerError, "failed to update client")
			return
		}
		writeJSONResponse(w, http.StatusOK, map[string]string{"status": "updated", "clientID": id})
	}
}

func (cs *ClientService) DeleteClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if err := cs.repo.DeleteClient(r.Context(), id); err != nil {
			writeJSONError(w, http.StatusInternalServerError, "failed to delete client")
			return
		}
		writeJSONResponse(w, http.StatusOK, map[string]string{"status": "deleted", "clientID": id})
	}
}

func writeJSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	writeJSONResponse(w, code, map[string]interface{}{
		"code":    code,
		"message": message,
	})
}
