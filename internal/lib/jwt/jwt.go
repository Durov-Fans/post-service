package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
)

func ValidateToken(tokenString string, secret string) (int64, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		log.Println(tokenString)
		return []byte(secret), nil
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
