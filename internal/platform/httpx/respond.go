package httpx

import (
	"encoding/json"
	"net/http"
)

// JSON escreve v como JSON com o status informado.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Error escreve {"error": msg} (mesmo formato dos serviços antigos).
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// Decode lê o corpo JSON (máx. 1 MiB). Em caso de erro já responde 400 e devolve false.
func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v); err != nil {
		Error(w, http.StatusBadRequest, "corpo da requisição inválido")
		return false
	}
	return true
}
