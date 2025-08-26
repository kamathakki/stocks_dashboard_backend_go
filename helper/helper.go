package helper

import (
	"encoding/json"
	"net/http"
	"stock_automation_backend_go/shared/env"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func WriteJson(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	jsonData, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

	w.Write(jsonData)
}

func ComparePassword(password, hashedPassword []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}

func CreateToken(payload struct {
	ID                           int64
	UserName, DisplayName, Email string
}, tokenType string) (string, error) {
	secret := []byte(env.GetEnv(env.EnvKeys.SECRET_KEY))
	exp := time.Now().Add(24 * time.Hour)

	if tokenType != "A" {
		secret = []byte(env.GetEnv(env.EnvKeys.REFRESH_SECRET_KEY))
		exp = time.Now().Add(7 * 24 * time.Hour)
	}

	claims := jwt.MapClaims{}
	claims["id"] = payload.ID
	claims["userName"] = payload.UserName
	claims["DisplayName"] = payload.DisplayName
	claims["Email"] = payload.Email
	claims["exp"] = exp.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)

}
