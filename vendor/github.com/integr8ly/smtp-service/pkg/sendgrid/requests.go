package sendgrid

import (
	"encoding/json"

	"github.com/pkg/errors"
)

func buildCreateSubUserBody(username, email, password string, ips []string) ([]byte, error) {
	body := struct {
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Password string   `json:"password"`
		IPs      []string `json:"ips"`
	}{
		Username: username,
		Email:    email,
		Password: password,
		IPs:      ips,
	}
	return marshalRequestBody(&body, "sub user create")
}

func buildCreateAPIKeyBody(id string, scopes []string) ([]byte, error) {
	body := struct {
		Name   string   `json:"name"`
		Scopes []string `json:"scopes"`
	}{
		Name:   id,
		Scopes: scopes,
	}
	return marshalRequestBody(&body, "api key create")
}

func marshalRequestBody(body interface{}, bodyDesc string) ([]byte, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to construct %s body", bodyDesc)
	}
	return bodyJSON, nil
}
