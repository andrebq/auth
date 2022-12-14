package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/andrebq/auth/internal/usererror"
	"github.com/dghubble/sling"
)

type (
	C struct {
		base *sling.Sling
	}
)

func New(base string) *C {
	if strings.HasSuffix(base, "//") {
		base = fmt.Sprintf("%v/", strings.ReplaceAll(base, "//", ""))
	}
	if !strings.HasSuffix(base, "/") {
		base = fmt.Sprintf("%v/", base)
	}
	return &C{
		base: sling.New().Base(base),
	}
}

func (c *C) Login(ctx context.Context, login, password string) error {
	var ue usererror.E
	var out interface{}
	res, err := c.base.BodyJSON(struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}{
		Login:    login,
		Password: password,
	}).Post("/auth/login").Receive(&out, &ue)
	if err != nil {
		return err
	} else if ue.Failure() {
		return ue
	} else if res.StatusCode != http.StatusOK {
		return fmt.Errorf("client: unexpected status code %v", res.StatusCode)
	}
	return nil
}

func (c *C) ValidateToken(ctx context.Context, token string) (string, string, error) {
	var ue usererror.E
	var out struct {
		UDI       string `json:"uid"`
		TokenID   string `json:"tokenID"`
		TokenType string `json:"tokenType"`
	}
	res, err := c.base.BodyJSON(struct {
		Token string `json:"token"`
	}{
		Token: token,
	}).Post("/auth/token").Receive(&out, &ue)
	if err != nil {
		return "", "", err
	} else if ue.Failure() {
		return "", "", ue
	} else if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("client: unexpected status code %v", res.StatusCode)
	}
	return out.UDI, out.TokenType, err
}

func (c *C) StartSession(ctx context.Context, login, password string, ttl time.Duration) (string, error) {
	var ue usererror.E
	var out struct {
		Token string `json:"token"`
	}
	res, err := c.base.BodyJSON(struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		TTL      string `json:"ttl"`
	}{
		Login:    login,
		Password: password,
		TTL:      ttl.String(),
	}).Post("/session").Receive(&out, &ue)
	if err != nil {
		return "", err
	} else if ue.Failure() {
		return "", ue
	} else if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("client: unexpected status code %v", res.StatusCode)
	}
	return out.Token, nil
}
