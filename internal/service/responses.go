package service

import (
	"encoding/json"
	"net/http"
)

func WriteJSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func WriteJSONError(w http.ResponseWriter, code int, message string) {
	WriteJSONResponse(w, code, map[string]interface{}{
		"code":    code,
		"message": message,
	})
}