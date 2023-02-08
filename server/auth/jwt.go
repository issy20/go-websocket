package auth

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/issy20/go-websocket/models"
)

const hmacSecret = "WjdwZUh2dWJGdFB1UWRybg=="
const defaultExpireTime = 604800 // 1 week

type Claims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	jwt.StandardClaims
}

func (c *Claims) GetId() string {
	return c.ID
}

func (c *Claims) GetName() string {
	return c.Name
}

func CreateJWTToken(user models.IUser) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"Id":        user.GetId(),
		"Name":      user.GetName(),
		"ExpiresAt": time.Now().Unix() + defaultExpireTime,
	})
	tokenString, err := token.SignedString([]byte(hmacSecret))
	return tokenString, err
}

func ValidateToken(tokenString string) (models.IUser, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(hmacSecret), nil
	})
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	} else {
		return nil, err
	}

}
