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

// Error escreve {"error": msg, "message": msg}. "error" é o formato dos serviços antigos;
// "message" é o campo que o frontend exibe ao usuário (err.error.message).
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg, "message": msg})
}

// Decode lê o corpo JSON (máx. 1 MiB). Em caso de erro já responde 400 e devolve false.
func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(v); err != nil {
		Error(w, http.StatusBadRequest, "corpo da requisição inválido")
		return false
	}
	return true
}
