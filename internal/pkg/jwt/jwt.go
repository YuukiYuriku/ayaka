package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"gitlab.com/ayaka/config"
	"gitlab.com/ayaka/internal/pkg/customerrors"
)

type JwtHandler struct {
	Config *config.Config `inject:"config"`
}

type Claims struct {
	UserCode string `json:"userCode"`
	UserName string `json:"userName"`
	jwt.RegisteredClaims
}

func (j *JwtHandler) GenerateToken(userCode, userName, signKey string, duration time.Duration, uot time.Duration) (string, error) {
	claims := &Claims{
		UserCode: userCode,
		UserName: userName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration * uot)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signKey))
}

func (j *JwtHandler) DecodeAndVerifyJWT(tokenString string, signKey []byte) (*Claims, error) {
	// Parse the token with claims.
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Ensure the signing method is correct.
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return signKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenSignatureInvalid) || errors.Is(err, jwt.ErrSignatureInvalid) {
			return nil, customerrors.ErrInvalidToken
		}
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	// Check if the token is valid and claims are of type `*Claims`.
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, customerrors.ErrInvalidToken
}
