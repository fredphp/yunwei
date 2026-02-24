package utils

import (
        "crypto/md5"
        "encoding/hex"
        "errors"
        "time"

        "github.com/golang-jwt/jwt/v5"
)

type CustomClaims struct {
        ID       uint   `json:"id"`
        Username string `json:"username"`
        Role     string `json:"role"`
        jwt.RegisteredClaims
}

func GenerateToken(id uint, username, role string) (string, error) {
        claims := CustomClaims{
                ID:       id,
                Username: username,
                Role:     role,
                RegisteredClaims: jwt.RegisteredClaims{
                        ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
                        IssuedAt:  jwt.NewNumericDate(time.Now()),
                        Issuer:    "yunwei",
                },
        }

        token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
        return token.SignedString([]byte("yunwei-secret-key"))
}

func ParseToken(tokenString string) (*CustomClaims, error) {
        token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
                return []byte("yunwei-secret-key"), nil
        })

        if err != nil {
                return nil, err
        }

        if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
                return claims, nil
        }

        return nil, errors.New("invalid token")
}
