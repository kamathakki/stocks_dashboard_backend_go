package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"stock_automation_backend_go/helper"
	"stock_automation_backend_go/shared/env"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func VerifyTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			helper.WriteJson(w, http.StatusOK, nil, errors.New("token not set"))
			return
		}

		const bearerPrefix = "Bearer "
		if !strings.HasPrefix(authHeader, bearerPrefix) {
			helper.WriteJson(w, http.StatusOK, nil, errors.New("token not found"))
			return
		}

		tokenString := strings.TrimPrefix(authHeader, bearerPrefix)
		secret := []byte(env.GetEnv[string](env.EnvKeys.SECRET_KEY))

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			helper.WriteJson(w, http.StatusOK, nil, errors.New("invalid or expired token"))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func VerifyRefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserData struct {
			ID          int64  `json:"id"`
			UserName    string `json:"userName"`
			DisplayName string `json:"displayName"`
			Email       string `json:"email"`
		} `json:"userData"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		helper.WriteJson(w, http.StatusOK, map[string]any{"isTokenValid": false}, errors.New("user data not set"))
		return
	}

	if body.UserData.UserName == "" && body.UserData.Email == "" && body.UserData.ID == 0 {
		helper.WriteJson(w, http.StatusOK, map[string]any{"isTokenValid": false}, errors.New("user data not set"))
		return
	}

	c, err := r.Cookie("jwt")
	if err != nil {
		helper.WriteJson(w, http.StatusOK, map[string]any{"isTokenValid": false}, errors.New("refresh token not set"))
		return
	}

	refreshToken := strings.TrimPrefix(c.Value, "Bearer ")
	if refreshToken == "" {
		helper.WriteJson(w, http.StatusOK, map[string]any{"isTokenValid": false}, errors.New("refresh token not found"))
		return
	}

	refreshSecret := []byte(env.GetEnv[string](env.EnvKeys.REFRESH_SECRET_KEY))
	token, parseErr := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return refreshSecret, nil
	})

	if parseErr != nil || !token.Valid {
		helper.WriteJson(w, http.StatusOK, map[string]any{"isTokenValid": false}, errors.New("invalid or expired refresh token"))
		return
	}

	newToken, createErr := helper.CreateToken(struct {
		ID                           int64
		UserName, DisplayName, Email string
	}{
		ID:          body.UserData.ID,
		UserName:    body.UserData.UserName,
		DisplayName: body.UserData.DisplayName,
		Email:       body.UserData.Email,
	}, "A")

	if createErr != nil {
		helper.WriteJson(w, http.StatusOK, map[string]any{"isTokenValid": false}, errors.New("failed to create access token"))
		return
	}

	helper.WriteJson(w, http.StatusOK, map[string]any{
		"token":        newToken,
		"isTokenValid": true,
	}, nil)
}
