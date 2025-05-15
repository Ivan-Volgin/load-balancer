package service

import (
	"encoding/json"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"load-balancer/internal/models"
	"load-balancer/internal/repo"
	"net/http"
	"time"
)

type ClientService struct {
	logger *zap.SugaredLogger
	repo   repo.Repository
}

func NewClientService(repo repo.Repository, logger *zap.SugaredLogger) *ClientService {
	return &ClientService{
		repo:   repo,
		logger: logger,
	}
}

func (cs *ClientService) CreateClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RateLimitClient
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid request body")
			cs.logger.Error(errors.Wrap(err, "invalid request body"))
			return
		}

		if req.ClientID == "" || req.Capacity <= 0 || req.RatePerSecond <= 0 {
			WriteJSONError(w, http.StatusBadRequest, "invalid client data")
			cs.logger.Error(errors.Wrap(err, "invalid client data"))
			return
		}

		req.Tokens = req.Capacity
		req.LastRefillAt = time.Now().Unix()

		if err := cs.repo.CreateClient(r.Context(), req); err != nil {
			WriteJSONError(w, http.StatusInternalServerError, "failed to create client")
			cs.logger.Error(errors.Wrap(err, "failed to create client"))
			return
		}

		WriteJSONResponse(w, http.StatusCreated, map[string]string{"status": "created", "clientID": req.ClientID})
		cs.logger.Infow("created client", "clientID", req.ClientID)
	}
}

func (cs *ClientService) GetClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		client, err := cs.repo.GetClientByID(r.Context(), id)
		if err != nil {
			WriteJSONError(w, http.StatusInternalServerError, "failed to get client")
			cs.logger.Error(errors.Wrap(err, "failed to get client"))
			return
		}
		if client == nil {
			WriteJSONError(w, http.StatusNotFound, "client not found")
			cs.logger.Error(errors.New("client not found"))
			return
		}
		WriteJSONResponse(w, http.StatusOK, client)
		cs.logger.Infow("client found", "client", client)
	}
}

func (cs *ClientService) UpdateClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")

		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			WriteJSONError(w, http.StatusBadRequest, "invalid request body")
			cs.logger.Error(errors.Wrap(err, "invalid request body"))
			return
		}

		if len(updates) == 0 {
			WriteJSONError(w, http.StatusBadRequest, "no fields to update")
			cs.logger.Error(errors.New("no fields to update"))
			return
		}

		existingClient, err := cs.repo.GetClientByID(r.Context(), id)
		if err != nil {
			WriteJSONError(w, http.StatusInternalServerError, "failed to get client")
			cs.logger.Error(errors.Wrap(err, "failed to get client"))
			return
		}
		if existingClient == nil {
			WriteJSONError(w, http.StatusNotFound, "client not found")
			cs.logger.Error(errors.New("client not found"))
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
			WriteJSONResponse(w, http.StatusInternalServerError, "failed to update client")
			cs.logger.Error(errors.Wrap(err, "failed to update client"))
			return
		}
		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "updated", "clientID": id})
		cs.logger.Infow("updated client", "clientID", id)
	}
}

func (cs *ClientService) DeleteClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := cs.repo.DeleteClient(r.Context(), id); err != nil {
			WriteJSONError(w, http.StatusInternalServerError, "failed to delete client")
			cs.logger.Error(errors.Wrap(err, "failed to delete client"))
			return
		}
		WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "deleted", "clientID": id})
		cs.logger.Infow("deleted client", "clientID", id)
	}
}
