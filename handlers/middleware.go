package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dhruv15803/echo-blog-app/storage"
	"github.com/golang-jwt/jwt/v5"
)

var (
	AuthUserId = "authUserId"
)

func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {

	// extract token from request cookie
	// decode jwt token with secret and extract payload
	// make sure that the expiration ("exp") is greather than time.Now().Unix()
	// attatch payload (userId) to the request context

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("auth_token")
		if err != nil {
			writeJSONError(w, "auth cookie not found", http.StatusBadRequest)
			return
		}

		tokenStr := cookie.Value

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			return []byte(JWT_SECRET), nil
		})
		if err != nil {
			log.Printf("failed to parse jwt token :- %v\n", err.Error())
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Println("invalid claims")
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		if !token.Valid {
			log.Println("invalid token")
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		tokenExpirationTimeUnix := int64(claims["exp"].(float64))
		if time.Now().Unix() > tokenExpirationTimeUnix {
			log.Println("token has expired")
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		userId := int(claims["sub"].(float64))

		// attatch this payload to request context

		ctx := context.WithValue(r.Context(), AuthUserId, userId)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

func (h *Handler) AdminMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		userId, ok := r.Context().Value(AuthUserId).(int)
		if !ok {
			writeJSONError(w, "internal server error", http.StatusInternalServerError)
			return
		}

		user, err := h.storage.GetUserById(userId)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeJSONError(w, "user not found", http.StatusBadRequest)
				return
			} else {
				log.Printf("failed to get user by id :- %v\n", err.Error())
				writeJSONError(w, "internal server error", http.StatusInternalServerError)
				return
			}
		}

		if user.Role != storage.AdminRole {
			writeJSONError(w, "user is not an admin", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
