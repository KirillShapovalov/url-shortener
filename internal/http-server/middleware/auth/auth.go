package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"url-shortener/internal/http-server/middleware/logger"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// BasicAuthWithJSON - базовая аутентификация с возвратом JSON при ошибке
func BasicAuthWithJSON(realm string, users map[string]string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.FromContext(r.Context())
			if log == nil {
				log = slog.New(slog.NewTextHandler(os.Stdout, nil)) // Фоллбэк, если логгер отсутствует
			}
			username, password, ok := r.BasicAuth()

			// Если авторизация отсутствует или невалидна
			if !ok || !validateUser(username, password, users) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				err := json.NewEncoder(w).Encode(ErrorResponse{Error: "invalid authorization credentials"})
				if err != nil {
					log.Error("failed to encode JSON response", slog.String("error", err.Error()))
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				log.Info("unauthorized access attempt", slog.String("username", username))
				return
			}

			// Если авторизация успешна, продолжаем обработку запроса
			next.ServeHTTP(w, r)
		})
	}
}

// validateUser - проверяет учетные данные
func validateUser(username, password string, users map[string]string) bool {
	if pass, exists := users[username]; exists && pass == password {
		return true
	}
	return false
}
