package api

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GoMudEngine/GoMud/internal/configs"
	"github.com/GoMudEngine/GoMud/internal/mudlog"
	"github.com/GoMudEngine/GoMud/internal/users"
	"github.com/golang-jwt/jwt/v5"
)

var jwtSigningKey []byte

// InitJWT derives a stable signing key from the server seed.
// Must be called after configs are loaded.
func InitJWT() {
	seed := configs.GetConfig().Server.Seed.String()
	if seed == "" {
		seed = "gomud-default-jwt-seed"
	}
	hash := sha256.Sum256([]byte("gomud-jwt-key:" + seed))
	jwtSigningKey = hash[:]
}

const tokenExpiry = 2 * time.Hour

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string   `json:"token"`
	User      userInfo `json:"user"`
	ExpiresAt string   `json:"expiresAt"`
}

type userInfo struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type jwtClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Username and password required")
		return
	}

	// Load user and verify credentials
	user, err := users.LoadUser(req.Username)
	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if !user.PasswordMatches(req.Password) {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	if user.Role == users.RoleUser {
		writeError(w, http.StatusForbidden, "Admin access required")
		return
	}

	// Generate JWT
	expiresAt := time.Now().Add(tokenExpiry)
	claims := jwtClaims{
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.UserId),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(jwtSigningKey)
	if err != nil {
		mudlog.Error("JWT signing error", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: tokenStr,
		User: userInfo{
			UserId:   user.UserId,
			Username: user.Username,
			Role:     user.Role,
		},
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	expiresAt := time.Now().Add(tokenExpiry)
	newClaims := jwtClaims{
		Username: claims.Username,
		Role:     claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   claims.Subject,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenStr, err := token.SignedString(jwtSigningKey)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token: tokenStr,
		User: userInfo{
			Username: claims.Username,
			Role:     claims.Role,
		},
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

func handleMe(w http.ResponseWriter, r *http.Request) {
	claims, ok := getClaimsFromContext(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	writeJSON(w, http.StatusOK, userInfo{
		Username: claims.Username,
		Role:     claims.Role,
	})
}
