package tokens_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/vexrina/cinemaLibrary/pkg/tokens"
	"github.com/vexrina/cinemaLibrary/pkg/types"
)

func TestCreateToken(t *testing.T) {
	username := "testuser"
	adminFlag := true

	tokenString, err := tokens.CreateToken(username, adminFlag)
	if err != nil {
		t.Fatalf("Error creating token: %v", err)
	}

	parsedToken, err := jwt.ParseWithClaims(tokenString, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return tokens.JWTKey, nil
	})
	if err != nil {
		t.Fatalf("Error parsing token: %v", err)
	}

	if _, ok := parsedToken.Method.(*jwt.SigningMethodHMAC); !ok {
		t.Fatalf("Unexpected signing method: %v", parsedToken.Header["alg"])
	}

	claims, ok := parsedToken.Claims.(*types.Claims)
	if !ok || !parsedToken.Valid || claims.Username != username || claims.Admin != adminFlag {
		t.Fatalf("Invalid token: %v", tokenString)
	}
}

func TestExtractTokenFromRequest(t *testing.T) {
    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer mytesttoken")

    token := tokens.ExtractTokenFromRequest(req)

    expectedToken := "mytesttoken"
    if token != expectedToken {
        t.Errorf("Expected token: %s, got: %s", expectedToken, token)
    }

    reqWithoutAuth := httptest.NewRequest("GET", "/test", nil)

    tokenFromEmptyReq := tokens.ExtractTokenFromRequest(reqWithoutAuth)

    if tokenFromEmptyReq != "" {
        t.Errorf("Expected empty token, got: %s", tokenFromEmptyReq)
    }

    reqWithInvalidToken := httptest.NewRequest("GET", "/test", nil)
    reqWithInvalidToken.Header.Set("Authorization", "InvalidFormatToken")

    tokenFromInvalidReq := tokens.ExtractTokenFromRequest(reqWithInvalidToken)

    if tokenFromInvalidReq != "" {
        t.Errorf("Expected empty token, got: %s", tokenFromInvalidReq)
    }
}

func TestParseToken(t *testing.T) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": "test_user",
    })

    tokenString, err := token.SignedString(tokens.JWTKey)
    if err != nil {
        t.Fatalf("Error signing token: %v", err)
    }

    t.Run("Valid token", func(t *testing.T) {
        parsedToken, err := tokens.ParseToken(tokenString)

        assert.NoError(t, err, "Unexpected error parsing token")
        assert.NotNil(t, parsedToken, "Parsed token is nil")
    })

    t.Run("Invalid token", func(t *testing.T) {
        invalidToken := "invalid_token_string"

        parsedToken, err := tokens.ParseToken(invalidToken)

        assert.Error(t, err, "Expected error parsing invalid token")
        assert.Nil(t, parsedToken, "Parsed token is not nil")
    })
}

func TestValidateToken_EmptyToken(t *testing.T) {
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	adminFlag, err := tokens.ValidateToken(rr, req)

	assert.False(t, adminFlag, "Expected admin flag to be false")
	assert.Error(t, err, "Expected error for empty token")
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status code to be Unauthorized")
}

func TestValidateToken_BadToken(t *testing.T) {
	fakeToken := "InvalidToken"

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", fakeToken)

	rr := httptest.NewRecorder()

	adminFlag, err := tokens.ValidateToken(rr, req)

	assert.False(t, adminFlag, "Expected admin flag to be false")
	assert.Error(t, err, "Expected error for bad token")
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status code to be Unauthorized")
}

func TestValidateToken_ValidToken(t *testing.T) {
    fakeToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "username": "test_user",
    })
	tokenString, err := fakeToken.SignedString(tokens.JWTKey)
    if err != nil {
        t.Fatalf("Error signing token: %v", err)
    }
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+tokenString)

	rr := httptest.NewRecorder()

	adminFlag, err := tokens.ValidateToken(rr, req)

	assert.False(t, adminFlag, "Expected admin flag to be true")
	assert.NoError(t, err, "Unexpected error")
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status code to be OK")
}
