package keycloak

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gitlab.indev.by/2fa/2fa_go_be/internal/config"
	"gitlab.indev.by/2fa/2fa_go_be/internal/domain/user"
	apperros "gitlab.indev.by/2fa/2fa_go_be/internal/errors"
	"gitlab.indev.by/2fa/2fa_go_be/internal/external"
)

type KeyCloak struct {
	config config.Config
}

var claims struct {
	Phone string `json:"phone_number"`
}

func NewKeyCloak(config config.Config) KeyCloak {
	return KeyCloak{
		config: config,
	}
}

func (kc KeyCloak) GetPhone(user user.User) (string, error) {
	operation := "GetPhone"

	form := url.Values{}
	form.Add("grant_type", "password")
	form.Add("client_id", "2fa")
	form.Add("username", user.Email)
	form.Add("client_secret", kc.config.KeycloakClientSecret)
	form.Add("password", user.Password)
	form.Add("scope", "phone")

	tokenResp, err := http.PostForm("https://auth.indev.by/realms/indev/protocol/openid-connect/token", form)
	if err != nil {
		return "", apperros.Wrap(operation, err)
	}
	defer external.SafeCloseBody(tokenResp)

	if tokenResp.StatusCode == http.StatusUnauthorized {
		return "", apperros.Wrap(operation, apperros.ErrUnauthorized)
	} else if tokenResp.StatusCode != http.StatusOK {
		return "", apperros.Wrap(operation, apperros.ErrUnknown)
	}

	var tokenBody struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenBody); err != nil {
		return "", apperros.Wrap(operation, fmt.Errorf("cannot decode token: "+err.Error()))
	}

	phone, err := getPhoneFromToken(tokenBody.AccessToken)
	if err != nil {
		return "", apperros.Wrap(operation, err)
	}

	return phone, nil
}

func getPhoneFromToken(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid JWT token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("cannot decode JWT payload: " + err.Error())
	}

	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("cannot parse JWT claims: " + err.Error())
	}

	return claims.Phone, nil
}
