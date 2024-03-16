package tokens

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

	"github.com/vexrina/cinemaLibrary/pkg/types"
)

var JWTKey = []byte("your_secret_key")

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
	return token.SignedString(JWTKey)
}

// exported only for test
func ExtractTokenFromRequest(r *http.Request) string {
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

// exported only for test
func ParseToken(tokenString string) (*jwt.Token, error) {
	return jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		} else {
			return JWTKey, nil
		}
	})
}

func ValidateToken(w http.ResponseWriter, r *http.Request) (bool, error) {
	// get token from request
	tokenString := ExtractTokenFromRequest(r)

	// token is not empty
	if tokenString == "" {
		http.Error(w, "Token doesnot exist", http.StatusUnauthorized)
		return false, errors.New("token doesnot exist")
	}

	token, err := ParseToken(tokenString)
	if err != nil || !token.Valid {
		http.Error(w, "Bad token or token expired", http.StatusUnauthorized)
		return false, errors.New("bad token or token expired")
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok {
		http.Error(w, "Can not retrieve claims from token", http.StatusUnauthorized)
		return false, errors.New("can not retrieve claims from token")
	}

	return claims.Admin, nil
}
