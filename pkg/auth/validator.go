package auth

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Username string `json:"preferred_username"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

func GetUserId(ctx *fiber.Ctx) string {
	return ctx.Locals("claims").(*Claims).Name
}
func GetUserEmail(ctx *fiber.Ctx) string {
	return ctx.Locals("claims").(*Claims).Email
}

func NewKeycloakJWTValidator(issuerUrl, clientId string) (func(*fiber.Ctx, string) (bool, error), error) {

	myClient := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	ctx := oidc.ClientContext(context.Background(), myClient)
	provider, err := oidc.NewProvider(ctx, issuerUrl)
	if err != nil {
		return nil, err
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientId,
	})
	return func(c *fiber.Ctx, key string) (bool, error) {
		var ctx = c.UserContext()
		_, err := verifier.Verify(ctx, key)
		if err != nil {
			log.Println("key verify error", key, err)
			return false, err
		}
		token, _ := jwt.ParseWithClaims(key, &Claims{},
			func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
					return nil, fmt.Errorf("Unexpected signing method: %v",
						token.Header["alg"])
				}
				return key, nil
			})
		c.Locals("claims", token.Claims)
		return true, nil
	}, nil
}

type KeycloakConfig struct {
	Realm        string
	IssuerURL    string
	ClientID     string
	ClientSecret string
}
