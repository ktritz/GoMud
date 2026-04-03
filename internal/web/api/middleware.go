package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/GoMudEngine/GoMud/internal/util"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const claimsKey contextKey = "jwtClaims"

// jwtAuth validates the JWT token in the Authorization header.
func jwtAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeError(w, http.StatusUnauthorized, "Invalid authorization format (use Bearer token)")
			return
		}

		tokenStr := parts[1]
		claims := &jwtClaims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSigningKey, nil
		})

		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		ctx := context.WithValue(r.Context(), claimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// getClaimsFromContext retrieves JWT claims from the request context.
func getClaimsFromContext(r *http.Request) (*jwtClaims, bool) {
	claims, ok := r.Context().Value(claimsKey).(*jwtClaims)
	return claims, ok
}

// withReadLock wraps a handler with a MUD read lock.
func withReadLock(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		util.LockMud()
		defer util.UnlockMud()
		next(w, r)
	}
}

// withWriteLock wraps a handler with a MUD write lock.
func withWriteLock(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		util.LockMud()
		defer util.UnlockMud()
		next(w, r)
	}
}

// cors adds CORS headers for the SPA dev server.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
