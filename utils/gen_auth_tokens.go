package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"nerajima.com/NeraJima/configs"
)

type claims struct {
	Type   string `json:"token_type"`
	UserId string `json:"userId"`
	jwt.RegisteredClaims
}

func GenAuthTokens(user_id string) (access, refresh string) {
	accessSecret, refreshSecret := configs.EnvTokenSecrets()

	accessExpTime := time.Now().Add(time.Hour * (24 * 30))       // 30 days
	refreshExpTime := time.Now().Add(time.Hour * (24 * 365 * 2)) // 2 Years

	accessClaims := claims{
		"access",
		user_id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshClaims := claims{
		"refresh",
		user_id,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)

	accessSigningKey := []byte(accessSecret)
	refreshSigningKey := []byte(refreshSecret)

	accessSigned, _ := accessToken.SignedString(accessSigningKey)
	refreshSigned, _ := refreshToken.SignedString(refreshSigningKey)

	return accessSigned, refreshSigned
}

func VerifyAccessToken(token string) (string, claims, error) {
	accessSecret, _ := configs.EnvTokenSecrets()
	var tokenBody claims

	_, err := jwt.ParseWithClaims(token, &tokenBody, func(t *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})

	if err != nil {
		v, _ := err.(*jwt.ValidationError)
		if v.Errors == jwt.ValidationErrorExpired { // if token is expired, gen new token
			newToken, newTokenBody := genNewAccessToken(tokenBody, accessSecret)
			return newToken, newTokenBody, nil
		} else {
			return "", claims{}, err
		}
	} else {
		timeInTwelveHours := time.Now().Add(time.Hour * 12).Unix()
		if timeInTwelveHours-tokenBody.ExpiresAt.Unix() > 0 { // if token will be expired within 12 hours, gen new token
			newToken, newTokenBody := genNewAccessToken(tokenBody, accessSecret)
			return newToken, newTokenBody, nil
		} else {
			return token, tokenBody, nil
		}
	}
}

func VerifyRefreshToken(token string) (string, claims, error) {
	_, refreshSecret := configs.EnvTokenSecrets()
	var tokenBody claims

	_, err := jwt.ParseWithClaims(token, &tokenBody, func(t *jwt.Token) (interface{}, error) {
		return []byte(refreshSecret), nil
	})

	if err != nil {
		return "", claims{}, err
	} else {
		return token, tokenBody, nil
	}
}

func VerifyAccessTokenNoRefresh(token string) (string, claims, error) {
	accessSecret, _ := configs.EnvTokenSecrets()
	var tokenBody claims

	_, err := jwt.ParseWithClaims(token, &tokenBody, func(t *jwt.Token) (interface{}, error) {
		return []byte(accessSecret), nil
	})

	if err != nil {
		return "", claims{}, err
	} else {
		return token, tokenBody, nil
	}
}

func genNewAccessToken(body claims, secret string) (string, claims) {
	accessExpTime := time.Now().Add(time.Hour * (24 * 30)) // 30 days

	body.IssuedAt = jwt.NewNumericDate(time.Now())
	body.ExpiresAt = jwt.NewNumericDate(accessExpTime)

	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, body)
	accessSigningKey := []byte(secret)
	newToken, _ := newAccessToken.SignedString(accessSigningKey)

	return newToken, body
}
