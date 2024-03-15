package tokens

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/vexrina/cinemaLibrary/pkg/types"
)

var jwtKey = []byte("your_secret_key")

func CreateToken(username string, adminflag bool) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &types.Claims{
		Username: username,
		Admin:    adminflag,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}


func extractTokenFromRequest(r *http.Request) string {
    // Извлекаем токен из хэдера Authorization
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return ""
    }

    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return ""
    }

    return parts[1]
}

func parseToken(tokenString string) (*jwt.Token, error) {
    return jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, jwt.ErrSignatureInvalid
        }
        return jwtKey, nil
    })
}

func ValidateToken(w http.ResponseWriter, r *http.Request) (bool, error) {
		// get token from request
		tokenString := extractTokenFromRequest(r)

		// token is not empty
		if tokenString == "" {
			http.Error(w, "Token doesnot exist", http.StatusUnauthorized)
			return false, errors.New("Token doesnot exist")
		}

		// check format token
		parts := strings.Split(tokenString, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Bad token", http.StatusUnauthorized)
			return false, errors.New("Bad token")
		}

		token, err:= parseToken(parts[1])
		if err != nil || !token.Valid {
			http.Error(w, "Bad token or token expired", http.StatusUnauthorized)
			return false, errors.New("Bad token or token expired")
		}

		claims, ok := token.Claims.(*types.Claims)
		if !ok{
			http.Error(w, "Can not retrieve claims from token", http.StatusUnauthorized)
			return false, errors.New("Can not retrieve claims from token")
		}
		
		return claims.Admin, nil
	}