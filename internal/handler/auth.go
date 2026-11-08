package handler

import (
	"context"
	"errors"
	"log"
	"net/http"
	"smartwallet/internal/service"

	"github.com/google/uuid"
)

type contextKey string

const userIDKey contextKey = "userID"

func AuthMiddleware(
	sessionService *service.SessionService,
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_token")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		token := cookie.Value

		userID, err := sessionService.ValidateToken(r.Context(), token)
		if err != nil {
			if errors.Is(err, service.ErrInvalidSession) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			log.Printf("validate session: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(userIDKey).(uuid.UUID)
	return userID, ok
}
