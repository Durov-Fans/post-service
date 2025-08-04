package jwt

import (
	"context"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
)

type JWT struct {
	secret string
}

func NewJWT(secret string) *JWT {
	return &JWT{secret}
}
func (s JWT) ValidateToken(tokenString string) (int64, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		log.Println(tokenString)
		return []byte(s.secret), nil
	})
	if err != nil || !token.Valid {
		log.Println(err)
		return 0, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}
	uid, ok := claims["userID"].(float64)
	if !ok {
		return 0, errors.New("user_id missing")
	}
	return int64(uid), nil
}
func (s JWT) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, "Missing or invalid token", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		userID, err := s.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userID)
		log.Println(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
