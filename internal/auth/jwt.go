package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var secret = []byte("secret")

func CreateJWT(username string) (string, error) {
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": username,
		"iss": "companies",
		"aud": "user",
		"exp": time.Now().Add(15 * time.Minute).Unix(),
		"iat": time.Now().Unix(),
	})

	tokenString, err := claims.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

//

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				log.Printf("missing token")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenParts := strings.Split(token, " ")
			if len(tokenParts) != 2 && tokenParts[0] != "Bearer" {
				log.Printf("invalid token parts")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			jwtToken, err := jwt.Parse(tokenParts[1], func(token *jwt.Token) (interface{}, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, fmt.Errorf("invalid method")
				}
				return []byte(secret), nil
			})
			if err != nil || !jwtToken.Valid {
				log.Printf("invalid token")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			claims, ok := jwtToken.Claims.(jwt.MapClaims)
			if !ok {
				log.Printf("invalid token claims")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
				log.Printf("expired token")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		},
	)
}
