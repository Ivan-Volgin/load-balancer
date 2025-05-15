package service

import (
	"encoding/json"
	"net/http"
)

// WriteJSONResponse — отправляет структурированный JSON-ответ клиенту.
func WriteJSONResponse(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteJSONError — отправляет JSON-ошибку с кодом состояния, чтобы унифицировать ответы от сервиса.
func WriteJSONError(w http.ResponseWriter, code int, message string) {
	WriteJSONResponse(w, code, map[string]interface{}{
		"code":    code,
		"message": message,
	})
}
