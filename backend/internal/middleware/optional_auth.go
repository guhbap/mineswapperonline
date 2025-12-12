package middleware

import (
	"context"
	"net/http"
	"strings"

	"minesweeperonline/internal/auth"
)

// OptionalAuthMiddleware проверяет токен, если он есть, но не требует его обязательного наличия
func OptionalAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]
				claims, err := auth.ValidateToken(tokenString)
				if err == nil {
					// Токен валиден, добавляем userID в контекст
					ctx := context.WithValue(r.Context(), "userID", claims.UserID)
					ctx = context.WithValue(ctx, "username", claims.Username)
					r = r.WithContext(ctx)
				}
				// Если токен невалиден, просто игнорируем и продолжаем без userID
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

