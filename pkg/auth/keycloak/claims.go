package keycloak

import "github.com/golang-jwt/jwt/v5"

var tokenMapClaims = defaultTokenMapClaims

func defaultTokenMapClaims(token *jwt.Token) (jwt.MapClaims, bool) {
	claims, ok := token.Claims.(jwt.MapClaims)
	return claims, ok
}
